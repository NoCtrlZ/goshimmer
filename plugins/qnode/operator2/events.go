package operator2

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// triggered by new notifyReqMsg, when another node notifies about
// its requests
func (op *scOperator) eventNotifyReqMsg(msg *notifyReqMsg) {
	log.Debugw("eventNotifyReqMsg",
		"num", len(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	stateIndex := op.stateTx.MustState().StateIndex()
	nextState := msg.StateIndex == stateIndex+1
	if !nextState && msg.StateIndex != stateIndex {
		log.Warn("request notification is not for current nor next state index. Ignored")
		return
	}
	// account notifications
	op.accountRequestIdNotifications(msg.SenderIndex, msg.StateIndex, msg.RequestIds...)
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
	op.accountRequestIdNotifications(op.PeerIndex(), op.stateTx.MustState().StateIndex(), req.reqId)

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
	log.Debugw("eventStartProcessingReqMsg",
		"req", msg.RequestId.Short(),
		"ts", msg.Timestamp,
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	stateIndex := op.stateTx.MustState().StateIndex()
	req, ok := op.requestFromId(msg.RequestId)
	if !ok {
		return // already processed
	}
	compReq := &computationRequest{
		req:             req,
		ts:              msg.Timestamp,
		leaderPeerIndex: msg.SenderIndex,
	}
	switch {
	case msg.StateIndex == stateIndex:
		op.currentStateCompRequests = append(op.currentStateCompRequests, compReq)
	case msg.StateIndex == stateIndex+1:
		op.nextStateCompRequests = append(op.nextStateCompRequests, compReq)
	default:
		return
	}
	op.takeAction()
}

// triggered by the signed result, sent by the the node to the leader
func (op *scOperator) eventSignedHashMsg(msg *signedHashMsg) {
	log.Debugf("eventSignedHashMsg")
	if op.leaderStatus == nil {
		log.Debugf("eventSignedHashMsg: op.leaderStatus == nil")
		// shouldn't be
		return
	}
	if msg.StateIndex != op.stateTx.MustState().StateIndex() {
		log.Debugf("eventSignedHashMsg: msg.StateIndex != op.stateTx.MustState().StateIndex()")
		return
	}
	if !msg.RequestId.Equal(op.leaderStatus.req.reqId) {
		log.Debugf("eventSignedHashMsg: !msg.RequestId.Equal(op.leaderStatus.req.reqId)")
		return
	}
	if !msg.OrigTimestamp.Equal(op.leaderStatus.ts) {
		log.Debugw("eventSignedHashMsg: !msg.OrigTimestamp.Equal(op.leaderStatus.ts)",
			"msgTs", msg.OrigTimestamp,
			"ownTs", op.leaderStatus.ts)
		return
	}
	if op.leaderStatus.signedHashes[msg.SenderIndex].MasterDataHash != nil {
		// repeating
		log.Debugf("eventSignedHashMsg: op.leaderStatus.signedHashes[msg.SenderIndex].MasterDataHash != nil")
		return
	}
	if req, ok := op.requestFromId(msg.RequestId); ok {
		req.log.Debugw("eventSignedHashMsg",
			"origTS", msg.OrigTimestamp,
			"stateIdx", msg.StateIndex,
		)
	}
	op.leaderStatus.signedHashes[msg.SenderIndex].MasterDataHash = msg.DataHash
	op.leaderStatus.signedHashes[msg.SenderIndex].SigBlocks = msg.SigBlocks
	op.takeAction()
}

// triggered from main msg queue whenever calculation of new result is finished

func (op *scOperator) eventResultCalculated(ctx *runtimeContext) {
	log.Debugf("eventResultCalculated")
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
	if ctx.leaderIndex == op.PeerIndex() {
		op.saveOwnResult(ctx)
	} else {
		op.sendResultToTheLeader(ctx)
	}
	op.takeAction()
}

func (op *scOperator) eventTimer(msg timerMsg) {
	if msg%50 == 0 {
		op.takeAction()
	}
	if msg%300 == 0 {
		log.Infof("eventTimer #%d", msg)
	}
}
