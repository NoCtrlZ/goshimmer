package commtypes

import "github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"

type StateTransactionMsg struct {
	*sctransaction.Transaction
}

type RequestMsg struct {
	*sctransaction.Transaction
	Index uint16
}
