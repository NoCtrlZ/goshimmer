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
	op.variableState = msg.VariableState
	op.stateTx = msg.StateTransaction

}

func (op *Operator) EventNotifyReqMsg(msg *committee.NotifyReqMsg) {

}

func (op *Operator) EventStartProcessingReqMsg(msg *committee.StartProcessingReqMsg) {

}

func (op *Operator) EventSignedHashMsg(msg *committee.SignedHashMsg) {

}

func (op *Operator) EventRequestMsg(msg committee.RequestMsg) {

}

func (op *Operator) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
