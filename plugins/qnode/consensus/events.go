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
func (op *Operator) EventRequestMsg(reqMsg *committee.RequestMsg) {
	if err := op.validateRequestBlock(reqMsg); err != nil {
		log.Warnw("request block validation failed.Ignored",
			"req", reqMsg.Id().Short(),
			"err", err,
		)
		return
	}
	req := op.requestFromMsg(reqMsg)
	req.log.Debugf("eventRequestMsg: id = %s", reqMsg.Id().Short())

	// include request into own list of the current state
	op.appendRequestIdNotifications(op.committee.OwnPeerIndex(), op.stateTx.MustState().StateIndex(), req.reqId)

	// the current leader is notified about new request
	op.sendRequestNotificationsToLeader([]*request{req})
	op.takeAction()
}

func (op *Operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {
	log.Debugw("EventNotifyReqMsg",
		"num", len(msg.RequestIds),
		"sender", msg.SenderIndex,
		"stateIdx", msg.StateIndex,
	)
	op.MustValidStateIndex(msg.StateIndex)

	// include all reqids into notifications list
	op.appendRequestIdNotifications(msg.SenderIndex, msg.StateIndex, msg.RequestIds...)
	op.takeAction()
}

func (op *Operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {
	log.Debugw("EventStartProcessingReqMsg",
		"reqId", msg.RequestId.Short(),
		"sender", msg.SenderIndex,
	)

	op.MustValidStateIndex(msg.StateIndex)

	// TODO

}

func (op *Operator) EventSignedHashMsg(msg *committee.SignedHashMsg) {
	log.Debugw("EventSignedHashMsg",
		"reqId", msg.RequestId.Short(),
		"sender", msg.SenderIndex,
	)

	// TODO

}

func (op *Operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
