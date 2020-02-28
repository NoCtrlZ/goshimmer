package vm

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

type RuntimeContext interface {
	InputVars() generic.ValueMap
	OutputVars() generic.ValueMap
	SetError(error)
	RequestTransferId() *hashing.HashValue
	GetDepositOutput() (uint16, uint64)
	GetRandom() uint32
}
