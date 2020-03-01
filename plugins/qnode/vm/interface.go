package vm

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

type RuntimeContext interface {
	RequestVars() generic.ValueMap
	StateVars() generic.ValueMap
	ConfigVars() generic.ValueMap
	SetError(error)
	Error() error
	MainRequestOutputs() [3]*generic.OutputRefWithAddrValue
	RequestTransferId() *HashValue
	Signature() []byte
	SendFundsToAddress([]*generic.OutputRef, *HashValue)
	AddRequestToSelf(uint16)
}
