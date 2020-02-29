package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
)

// triggered by new request msg from the node
// called from he main queue

func (op *AssemblyOperator) EventRequestMsg(reqRef *sc.RequestRef) {
	if err := op.validateRequest(reqRef); err != nil {
		log.Errorw("invalid request message received",
			"req", reqRef.Id().Short(),
			"err", err,
		)
		return
	}
	reqRec := op.requestFromMsg(reqRef)
	log.Debugw("EventRequestMsg",
		"tx", reqRef.Tx().ShortStr(),
		"reqIdx", reqRef.Index(),
		"req id", reqRef.Id().Short(),
		"leader", op.currentLeaderIndex(reqRec),
		"iAmTheLeader", op.iAmCurrentLeader(reqRec),
	)
	op.takeAction()
}

// triggered by the new stateTx update

func (op *AssemblyOperator) EventStateUpdate(tx sc.Transaction) {
	log.Debugw("EventStateUpdate", "tx", tx.ShortStr())

	stateUpd := tx.MustState()
	if stateUpd.Error() != nil {
		log.Warnf("Ignore the state update with error: '%v'", stateUpd.Error())
		return
	}

	state := op.stateTx.MustState()

	if stateUpd.StateIndex() <= state.StateIndex() {
		// wrong sequence of stateTx indices. Ignore the message
		log.Warnf("wrong sequence of stateTx indices. Ignore the message")
		return
	}
	reqId := stateUpd.RequestId()
	req, ok := op.requestFromId(reqId)
	if !ok {
		// already processed
		return
	}
	duration := time.Duration(0)
	if req.reqRef != nil {
		duration = time.Since(req.whenMsgReceived)
	}
	log.Infow("RECEIVE STATE UPD",
		"stateIndex", stateUpd.StateIndex(),
		"tx", tx.ShortStr(),
		"req", reqId.Short(),
		"duration", duration)

	// delete processed request from buffer
	op.markRequestProcessed(reqId, duration)

	if !state.Config().Id().Equal(stateUpd.Config().Id()) {
		// configuration changed
		ownAddr, ownPort := op.comm.GetOwnAddressAndPort()
		iAmParticipant, err := op.configure(stateUpd.Config().Id(), ownAddr, ownPort)
		if err != nil || !iAmParticipant {
			op.Dismiss()
			return
		}
	}
	// update current state
	op.stateTx = tx
	op.adjustToContext()
	op.takeAction()
}

// triggered from main msg queue whenever calculation of new result is finished

func (op *AssemblyOperator) EventResultCalculated(ctx *runtimeContext) {
	reqId := ctx.reqRef.Id()
	reqRec, ok := op.requestFromId(reqId)
	if !ok {
		// processed
		return
	}
	log.Debugw("EventResultCalculated",
		"req id", reqId.Short(),
		"state idx", ctx.state.MustState().StateIndex(),
		"current state idx", op.stateTx.MustState().StateIndex(),
	)

	taskId := hashing.HashData(reqId.Bytes(), ctx.state.Id().Bytes())
	delete(reqRec.startedCalculation, *taskId)

	if ctx.err != nil {
		var err error
		ctx.resultTx, err = clientapi.ErrorTransaction(ctx.reqRef, ctx.state.MustState().Config(), ctx.err)
		if err != nil {
			log.Errorw("EventResultCalculated: error while processing error state",
				"req id", reqId.Short(),
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
	log.Debugw("EventResultCalculated in context", "req", ctx.reqRef.Id().Short())

	if reqRec.ownResultCalculated != nil {
		// shouldn't be
		if op.resultBelongsToContext(reqRec.ownResultCalculated.res) {
			panic("inconsistency: duplicate result")
		}
		// dismiss new result, which is from another R,E,S context
		return
	}
	// new result
	err := sc.SignTransaction(ctx.resultTx, op)
	if err != nil {
		log.Errorf("SignTransaction returned: %v", err)
		return
	}
	masterDataHash := ctx.resultTx.MasterDataHash()
	reqRec.ownResultCalculated = &resultCalculated{
		res:            ctx,
		resultHash:     resultHash(ctx.state.MustState().StateIndex(), reqId, masterDataHash),
		masterDataHash: masterDataHash,
	}
	op.takeAction()
}

// triggered by new result hash received from another operator
// called from the main queue

func (op *AssemblyOperator) EventPushResultMsg(pushMsg *pushResultMsg) {
	log.Debugw("EventPushResultMsg received",
		"from", pushMsg.SenderIndex,
		"req id", pushMsg.RequestId.Short(),
		"state idx", pushMsg.StateIndex,
	)
	if _, ok := op.requestFromId(pushMsg.RequestId); !ok {
		return // already processed
	}
	if err := op.accountNewPushMsg(pushMsg); err != nil {
		log.Errorf("accountNewPushMsg returned: %v", err)
		return
	}
	op.adjustToContext()
	op.takeAction()
}

func (op *AssemblyOperator) EventPullMsgReceived(msg *pullResultMsg) {
	log.Debug("EventPullResultMsg")
	req, ok := op.requestFromId(msg.RequestId)
	if !ok {
		return // already processed
	}
	req.pullMessages[msg.SenderIndex] = msg
	op.takeAction()
}

func (op *AssemblyOperator) EventTimer(msg timerMsg) {
	if msg%300 == 0 {
		log.Infow("EventTimer", "#", int(msg))
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
