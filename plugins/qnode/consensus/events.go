package consensus

import "github.com/iotaledger/goshimmer/plugins/qnode/commtypes"

func (op *ConsensusOperator) EventStateTransitionMsg(msg *commtypes.StateTransitionMsg) {

}

func (op *ConsensusOperator) EventNotifyReqMsg(msg *commtypes.NotifyReqMsg) {

}

func (op *ConsensusOperator) EventStartProcessingReqMsg(msg *commtypes.StartProcessingReqMsg) {

}

func (op *ConsensusOperator) EventSignedHashMsg(msg *commtypes.SignedHashMsg) {

}

func (op *ConsensusOperator) EventRequestMsg(msg commtypes.RequestMsg) {

}

func (op *ConsensusOperator) EventTimerMsg(msg commtypes.TimerTick) {
	if msg%10 == 0 {
		op.takeAction()
	}
}
