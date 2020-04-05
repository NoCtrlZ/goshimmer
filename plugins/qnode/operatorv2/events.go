package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// triggered by new notifyReqMsg
func (op *scOperator) eventNotifyReqMsg(msg *notifyReqMsg) {
	if msg.StateIndex < op.stateTx.MustState().StateIndex() {
		// ignore from state indices in the past
		return
	}
	if msg.Renew {
		op.requestNotificationsReceived[msg.SenderIndex][msg.StateIndex] = msg.RequestIds
	} else {
		op.requestNotificationsReceived[msg.SenderIndex][msg.StateIndex] =
			append(op.requestNotificationsReceived[msg.SenderIndex][msg.StateIndex], msg.RequestIds...)
		// may cause duplicates
	}
}

// triggered by new request msg from the node
// called from he main queue
func (op *scOperator) eventRequestMsg(reqRef *sc.RequestRef) {
	if err := op.validateRequestBlock(reqRef); err != nil {
		log.Errorw("invalid request message received. Ignored...",
			"req", reqRef.Id().Short(),
			"err", err,
		)
		return
	}
	reqRec := op.requestFromMsg(reqRef)
	reqRec.log.Debugw("eventRequestMsg",
		"tx", reqRef.Tx().ShortStr(),
		"reqIdx", reqRef.Index(),
		"leader", op.currentLeaderIndex(reqRec),
		"iAmTheLeader", op.iAmCurrentLeader(reqRec),
	)
	op.takeAction()
}

// triggered by the new stateTx update

func (op *scOperator) eventStateUpdate(tx sc.Transaction) {
	log.Debugw("eventStateUpdate", "tx", tx.ShortStr())

	stateUpd := tx.MustState()

	// current state is always present
	state := op.stateTx.MustState()

	if stateUpd.Error() == nil && stateUpd.StateIndex() <= state.StateIndex() {
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
		op.stateTx = tx
	} else {
		log.Warnf("state update with error ignored: '%v'", stateUpd.Error())
	}
	op.adjustToContext()
	op.takeAction()
}

// triggered from main msg queue whenever calculation of new result is finished

func (op *scOperator) eventResultCalculated(ctx *runtimeContext) {
	reqId := ctx.reqRef.Id()
	reqRec, ok := op.requestFromId(reqId)
	if !ok {
		// processed
		return
	}
	reqRec.log.Debugw("eventResultCalculated",
		"state idx", ctx.state.MustState().StateIndex(),
		"cur state idx", op.stateTx.MustState().StateIndex(),
		"resultErr", ctx.err,
	)

	taskId := hashing.HashData(reqId.Bytes(), ctx.state.Id().Bytes())
	delete(reqRec.startedCalculation, *taskId)

	if ctx.err != nil {
		var err error
		ctx.resultTx, err = clientapi.ErrorTransaction(ctx.reqRef, ctx.state.MustState().Config(), ctx.err)
		if err != nil {
			reqRec.log.Errorw("eventResultCalculated: error while processing error state",
				"state idx", ctx.state.MustState().StateIndex(),
				"current state idx", op.stateTx.MustState().StateIndex(),
				"error", err,
			)
			return
		}
	}
	if !op.resultBelongsToContext(ctx) {
		// stateTx changed while it was being calculated
		// dismiss the result
		return
	}

	if reqRec.ownResultCalculated != nil {
		// shouldn't be
		if op.resultBelongsToContext(reqRec.ownResultCalculated.res) {
			panic("inconsistency: duplicate result")
		}
		// dismiss new result, which is from another R,E,S context
		return
	}
	// new result
	err := sc.SignTransaction(ctx.resultTx, op.keyPool())
	if err != nil {
		reqRec.log.Errorf("SignTransaction returned: %v", err)
		return
	}
	masterDataHash := ctx.resultTx.MasterDataHash()
	reqRec.log.Debugw("eventResultCalculated:",
		"input tx", ctx.state.Id().Short(),
		"res tx", ctx.resultTx.Id().Short(),
		"master result hash", masterDataHash.Short(),
		"err", err,
	)
	reqRec.ownResultCalculated = &resultCalculated{
		res:            ctx,
		resultHash:     resultHash(ctx.state.MustState().StateIndex(), reqId, masterDataHash),
		masterDataHash: masterDataHash,
	}
	op.takeAction()
}

// triggered by new result hash received from another operator

func (op *scOperator) eventPushResultMsg(pushMsg *pushResultMsg) {
	req, ok := op.requestFromId(pushMsg.RequestId)
	req.msgCounter++
	if !ok {
		return // already processed, ignore
	}
	req.log.Debugf("eventPushResultMsg received from peer %d", pushMsg.SenderIndex)

	op.accountNewPushMsg(pushMsg)
	op.adjustToContext()
	op.takeAction()
}

func (op *scOperator) eventPullMsgReceived(msg *pullResultMsg) {
	req, ok := op.requestFromId(msg.RequestId)
	if !ok {
		return // already processed
	}
	req.msgCounter++
	req.log.Debug("EventPullResultMsg")
	req.pullMessages[msg.SenderIndex] = msg
	op.adjustToContext()
	op.takeAction()
}

func (op *scOperator) eventTimer(msg msg.timerMsg) {
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
