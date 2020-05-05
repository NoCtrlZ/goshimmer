package committee

import "github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"

type StateTransactionMsg struct {
	*sctransaction.Transaction
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index uint16
}

func (reqMsg *RequestMsg) Id() *sctransaction.RequestId {
	ret := sctransaction.NewRequestId(reqMsg.Transaction.Id(), reqMsg.Index)
	return &ret
}
