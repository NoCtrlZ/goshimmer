package consensus

import "github.com/iotaledger/goshimmer/plugins/qnode/commtypes"

func (op *Operator) EventStateTransitionMsg(msg *commtypes.StateTransitionMsg) {

}

func (op *Operator) EventNotifyReqMsg(msg *commtypes.NotifyReqMsg) {

}

func (op *Operator) EventStartProcessingReqMsg(msg *commtypes.StartProcessingReqMsg) {

}

func (op *Operator) EventSignedHashMsg(msg *commtypes.SignedHashMsg) {

}

func (op *Operator) EventRequestMsg(msg commtypes.RequestMsg) {

}

func (op *Operator) EventTimerMsg(msg commtypes.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
