package operator2

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// triggered by new notifyReqMsg, when another node notifies about
// its requests
func (op *scOperator) eventNotifyReqMsg(msg *notifyReqMsg) {
	stateIndex := op.stateTx.MustState().StateIndex()
	nextState := msg.StateIndex == stateIndex+1
	if !nextState && msg.StateIndex != stateIndex {
		log.Warn("request notification is not for current nor next state index. Ignored")
		return
	}
	// account notifications
	op.accountRequestIdNotifications(msg.SenderIndex, nextState, msg.RequestIds...)

	op.takeAction()
}

// triggered by new request msg from the node
func (op *scOperator) eventRequestMsg(reqRef *sc.RequestRef) {
	if err := op.validateRequestBlock(reqRef); err != nil {
		log.Errorw("invalid request message received. Ignored...",
			"req", reqRef.Id().Short(),
			"err", err,
		)
		return
	}
	req := op.requestFromMsg(reqRef)
	req.log.Debugw("eventRequestMsg", "id", reqRef.Id().Short())

	// include request in own list of the current state
	op.accountRequestIdNotifications(op.PeerIndex(), false, req.reqId)

	// the current leader is notified about new request
	op.sendRequestNotification(req)

	op.takeAction()
}

// triggered by the new stateTx update

func (op *scOperator) eventStateUpdate(tx sc.Transaction) {
	log.Debugw("eventStateUpdate", "tx", tx.ShortStr())

	stateUpd := tx.MustState()

	// current state is always present
	state := op.stateTx.MustState()

	if stateUpd.Error() == nil && stateUpd.StateIndex() != state.StateIndex()+1 {
		// wrong sequence of stateTx indices. Ignore the message
		log.Warnf("wrong sequence of stateTx indices. Ignore the message")
		return
	}
	reqRef, reqExists := stateUpd.RequestRef()
	if reqExists {
		reqId := reqRef.Id()
		req, ok := op.requestFromId(reqId)
		if !ok {
			// already processed
			return
		}
		// delete processed request from pending queue
		op.markRequestProcessed(req)
	}
	log.Debugw("RECEIVE STATE UPD",
		"stateIdx", stateUpd.StateIndex(),
		"tx", tx.ShortStr(),
		"err", stateUpd.Error(),
	)

	if stateUpd.Error() == nil {
		if !state.Config().Id().Equal(stateUpd.Config().Id()) {
			// configuration changed
			iAmParticipant, err := op.configure(stateUpd.Config().Id())
			if err != nil || !iAmParticipant {
				op.dismiss()
				return
			}
		}
		// update current state
		log.Infof("STATE CHANGE %d --> %d", state.StateIndex(), stateUpd.StateIndex())

		op.setNewState(tx)
	} else {
		log.Warnf("state update with error ignored: '%v'", stateUpd.Error())
	}
	op.takeAction()
}

// triggered by `startProcessingReq` message sent from the leader
// if timestamp is acceptable and the msg context is from the current state or the next
// include the message into the state
func (op *scOperator) eventStartProcessingReqMsg(msg *startProcessingReqMsg) {
	stateIndex := op.stateTx.MustState().StateIndex()
	var pos int
	switch {
	case msg.StateIndex == stateIndex:
		// current state
		pos = 0
	case msg.StateIndex == stateIndex+1:
		// next state
		pos = 1
	default:
		log.Warnf("ignore 'initReq' message for %s: state index is out of context", msg.RequestId.Short())
		return
	}
	if op.requestToProcess[pos][msg.SenderIndex].reqId != nil {
		log.Errorf("repeating 'processReq' message from peer %d", msg.SenderIndex)
		return
	}
	if op.requestToProcess[pos][msg.SenderIndex].req != nil {
		log.Panicf("can't be: op.requestToProcess[pos][msg.SenderIndex].req != nil")
	}
	op.requestToProcess[pos][msg.SenderIndex].reqId = msg.RequestId
	op.requestToProcess[pos][msg.SenderIndex].ts = msg.Timestamp

	if pos != 0 {
		// if not current state, do nothing
		return
	}
	if req, ok := op.requestFromId(msg.RequestId); ok && req.reqRef != nil {
		op.requestToProcess[0][msg.SenderIndex].req = req
		go op.processRequest(msg.SenderIndex)
	}
}

// triggered by the signed result, sent by the the node to the leader
func (op *scOperator) eventSignedHashMsg(msg *signedHashMsg) {
	// validate
	req := op.requestToProcess[0][op.PeerIndex()].req
	if op.requestToProcess[0][op.PeerIndex()].reqId == nil || req == nil {
		log.Errorf("eventSignedHashMsg: wrong state of the leader ")
		return
	}
	if op.requestToProcess[0][msg.SenderIndex].MasterDataHash != nil ||
		op.requestToProcess[0][msg.SenderIndex].SigBlocks != nil {
		log.Errorf("eventSignedHashMsg: wrong state of the peer %d", msg.SenderIndex)
		return
	}
	req.log.Debugw("eventSignedHashMsg",
		"senderIndex", msg.SenderIndex,
		"stateIndex", msg.StateIndex)
	op.requestToProcess[0][msg.SenderIndex].MasterDataHash = msg.DataHash
	op.requestToProcess[0][msg.SenderIndex].SigBlocks = msg.SigBlocks
	// do not check master hash because at this point own calculations may not be finished yet
}

// triggered from main msg queue whenever calculation of new result is finished

func (op *scOperator) eventResultCalculated(ctx *runtimeContext) {
	// check if result belongs to context
	if ctx.state.MustState().StateIndex() != op.stateTx.MustState().StateIndex() {
		// out of context. ignore
		return
	}
	reqId := ctx.reqRef.Id()
	req, ok := op.requestFromId(reqId)
	if !ok {
		// processed
		return
	}
	req.log.Debugw("eventResultCalculated",
		"state idx", ctx.state.MustState().StateIndex(),
		"cur state idx", op.stateTx.MustState().StateIndex(),
		"resultErr", ctx.err,
	)

	if ctx.err != nil {
		var err error
		ctx.resultTx, err = clientapi.ErrorTransaction(ctx.reqRef, ctx.state.MustState().Config(), ctx.err)
		if err != nil {
			req.log.Errorw("eventResultCalculated: error while processing error state",
				"state idx", ctx.state.MustState().StateIndex(),
				"current state idx", op.stateTx.MustState().StateIndex(),
				"error", err,
			)
			return
		}
	}
	req.log.Debugw("eventResultCalculated:",
		"input tx", ctx.state.Id().Short(),
		"res tx", ctx.resultTx.Id().Short(),
	)
	err := sc.SignTransaction(ctx.resultTx, op.keyPool())
	if err != nil {
		req.log.Errorf("SignTransaction returned: %v", err)
		return
	}
	sigs, err := ctx.resultTx.Signatures()
	if err != nil {
		req.log.Errorf("Signatures returned: %v", err)
		return
	}
	op.requestToProcess[ctx.leaderIndex][0].ownResult = ctx.resultTx
	op.requestToProcess[ctx.leaderIndex][0].MasterDataHash = ctx.resultTx.MasterDataHash()
	op.requestToProcess[ctx.leaderIndex][0].SigBlocks = sigs

	if ctx.leaderIndex != op.PeerIndex() {
		// send result hash and signatures to the leader
		op.sendResultToTheLeader(ctx.leaderIndex)
	}
	op.takeAction()
}

func (op *scOperator) eventTimer(msg timerMsg) {
	if msg%300 == 0 {
		log.Debugw("eventTimer", "#", int(msg))
		snap := op.getStateSnapshot()
		log.Debugf("%+v", snap)
	}
	if msg%300 == 0 {
		err := op.consistentState()
		if err != nil {
			log.Panicf("inconsistent stateTx: %v", err)
		}
	}
	if msg%50 == 0 {
		op.takeAction()
	}
}
