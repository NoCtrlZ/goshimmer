package vm

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

type RuntimeContext interface {
	SContractAccount() *HashValue
	Time() time.Time
	PrevTime() time.Time
	RequestVars() generic.ValueMap
	StateVars() generic.ValueMap
	SetError(error)
	Error() error
	MainRequestOutputs() sc.MainRequestOutputs
	MainInputAddress() (*HashValue, error)
	RequestTransferId() *HashValue
	Signature() []byte
	SendFundsToAddress([]*generic.OutputRef, *HashValue)
	SendOutputsToOutputs([]*generic.OutputRef, []value.Output, *HashValue) error
	AddRequestToSelf(uint16) error
	Log() *logger.Logger
}
