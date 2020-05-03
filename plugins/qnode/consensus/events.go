package consensus

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
)

func (op *Operator) EventStateTransitionMsg(msg *committee.StateTransitionMsg) {
	if op.variableState != nil {
		if !(op.variableState.StateIndex()+1 == msg.VariableState.StateIndex()) {
			panic("assertion failed: op.variableState.StateIndex()+1 == msg.VariableState.StateIndex()")
		}
	}
	op.setNewState(msg.StateTransaction, msg.VariableState)

	// TODO
}

// triggered by new request msg from the node
func (op *Operator) EventRequestMsg(reqMsg committee.RequestMsg) {
	if err := op.validateRequestBlock(&reqMsg); err != nil {
		log.Errorw("invalid request message received. Ignored...",
			"req", reqMsg.Id().Short(),
			"err", err,
		)
		return
	}
	req := op.requestFromMsg(&reqMsg)
	req.log.Debugw("eventRequestMsg", "id", reqMsg.Id().Short())

	// include request in own list of the current state
	op.accountRequestIdNotifications(op.committee.OwnPeerIndex(), op.stateTx.MustState().StateIndex(), req.reqId)

	// the current leader is notified about new request
	op.sendRequestNotificationsToLeader([]*request{req})
	op.takeAction()
}

func (op *Operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {

}

func (op *Operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {

}

func (op *Operator) EventSignedHashMsg(msg *committee.SignedHashMsg) {

}

func (op *Operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
