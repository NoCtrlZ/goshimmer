package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
)

// triggered by new request msg from the node
// called from he main queue

func (op *AssemblyOperator) EventRequestMsg(tx sc.Transaction, reqIndex uint16) {
	log.Debugw("EventRequestMsg", "tx", tx.ShortStr(), "reqIdx", reqIndex)
	op.requestFromMsg(tx, reqIndex)
	op.takeAction()
}

// triggered by the new stateTx update

func (op *AssemblyOperator) EventStateUpdate(tx sc.Transaction) {
	log.Debugw("EventStateUpdate", "tx", tx.ShortStr())

	stateUpd := tx.MustState()
	state := op.stateTx.MustState()

	if stateUpd.StateIndex() <= state.StateIndex() {
		// wrong sequence of stateTx indices. Ignore the message
		return
	}
	reqId := RequestId(stateUpd.RequestRef())
	req, _ := op.requestFromIdHash(reqId)
	duration := "unknown"
	if req.msgTx != nil {
		duration = fmt.Sprintf("%v", time.Since(req.whenMsgReceived))
	}
	log.Infow("RECEIVE STATE UPD",
		"peer", op.peerIndex(), "tx", tx.ShortStr(), "duration", duration)

	// delete request from queue
	delete(op.requests, *reqId)
	op.processedCounter++

	if !state.ConfigId().Equal(stateUpd.ConfigId()) {
		// configuration changed
		ownAddr, ownPort := op.comm.GetOwnAddressAndPort()
		iAmParticipant, err := op.configure(stateUpd.ConfigId(), ownAddr, ownPort)
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

func (op *AssemblyOperator) EventResultCalculated(res *resultCalculated) {
	log.Debug("EventResultCalculated")
	reqHash := RequestId(res.requestTx.Id(), res.requestIndex)
	reqRec, _ := op.requestFromIdHash(reqHash)

	rsHash := hashing.HashData(reqHash.Bytes(), res.state.Id().Bytes())
	delete(reqRec.startedCalculation, *rsHash)

	if !op.resultBelongsToContext(res) {
		// stateTx changed while it was being calculated
		// dismiss the result
		return
	}
	log.Debugw("EventResultCalculated (in context)", "req id", res.requestTx.Id().Short())

	if reqRec.ownResultCalculated != nil {
		// shouldn't be
		if op.resultBelongsToContext(reqRec.ownResultCalculated.res) {
			panic("inconsistency: duplicate result")
		}
		// dismiss new result, which is from another R,E,S context
		return
	}
	// new result
	err := op.signTheResultTx(res.resultTx)
	if err != nil {
		log.Errorf("signTheResultTx returned: %v", err)
		return
	}
	reqRec.ownResultCalculated = &resultCalculatedIntern{
		res:            res,
		rsHash:         rsHash,
		masterDataHash: res.resultTx.MasterDataHash(),
	}
	op.takeAction()
}

// triggered by new result hash received from another operator
// called from the main queue

func (op *AssemblyOperator) EventPushResultMsg(resHashMsg *pushResultMsg) {
	log.Debug("EventPushResultMsg")
	if err := op.accountNewResultHash(resHashMsg); err != nil {
		log.Errorf("accountNewResultHash returned: %v", err)
		return
	}
	op.adjustToContext()
	op.takeAction()
}

func (op *AssemblyOperator) EventPullMsgReceived(msg *pullResultMsg) {
	log.Debug("EventPullResultMsg")
	req, _ := op.requestFromIdHash(msg.RequestId)
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
