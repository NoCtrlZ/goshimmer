package vm

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

type Processor interface {
	Run(inputs VMInputs) VMOutput
}

type VMInputs interface {
	RequestMsg() *committee.RequestMsg
	Timestamp() time.Time
	StateTransaction() *sctransaction.Transaction
	VariableState() state.VariableState
	Log() *logger.Logger
}

type VMOutput struct {
	Inputs            VMInputs
	ResultTransaction *sctransaction.Transaction
	StateUpdate       state.StateUpdate
	Error             string
}
