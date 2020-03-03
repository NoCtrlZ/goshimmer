package vm

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/hive.go/logger"
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
	MainRequestOutputs() sc.MainRequestOutputs
	RequestTransferId() *HashValue
	Signature() []byte
	SendFundsToAddress([]*generic.OutputRef, *HashValue)
	AddRequestToSelf(uint16)
	Log() *logger.Logger
}
