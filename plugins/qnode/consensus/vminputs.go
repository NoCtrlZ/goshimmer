package consensus

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

// implements VMInputs interface
type runtimeContext struct {
	// input. Leader where results must be sent
	leaderPeerIndex uint16
	// input. Request
	reqMsg *committee.RequestMsg
	// timestamp of the context. Imposed by the leader
	timestamp time.Time
	// current state represented by the stateTx and variableState
	stateTx       *sctransaction.Transaction
	variableState state.VariableState
	// output of the computation, represented by the resultTx and stateUpdate
	log *logger.Logger
}

func (ctx *runtimeContext) RequestMsg() *committee.RequestMsg {
	return ctx.reqMsg
}

func (ctx *runtimeContext) Timestamp() time.Time {
	return ctx.timestamp
}

func (ctx *runtimeContext) StateTransaction() *sctransaction.Transaction {
	return ctx.stateTx
}

func (ctx *runtimeContext) VariableState() state.VariableState {
	return ctx.variableState
}

func (ctx *runtimeContext) Log() *logger.Logger {
	return ctx.log
}
