package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

type resultCalculated struct {
	requestTx    sc.Transaction
	requestIndex uint16
	state        sc.Transaction
	resultTx     sc.Transaction
}

func (ctx *resultCalculated) InputVars() generic.ValueMap {
	return ctx.requestTx.Requests()[ctx.requestIndex].Vars()
}

func (ctx *resultCalculated) OutputVars() generic.ValueMap {
	return ctx.resultTx.MustState().StateVars()
}

func newResultCalculated(reqTx sc.Transaction, reqIdx uint16, prevStateTx sc.Transaction) (*resultCalculated, error) {
	reqBlock := reqTx.Requests()[reqIdx]
	if reqBlock.IsConfigUpdateReq() {
		return nil, fmt.Errorf("req.IsConfigUpdateReq()")
	}
	state, _ := prevStateTx.State()

	if !reqBlock.AssemblyId().Equal(state.AssemblyId()) {
		return nil, fmt.Errorf("!req.AssemblyId().Equal(state.AssemblyId())")
	}

	nextStateVars := state.StateVars().Clone()

	tx := sc.NewTransaction()
	tr := tx.Transfer()
	// add request chain link
	tr.AddInput(value.NewInput(reqTx.Transfer().Id(), reqBlock.RequestChainOutputIndex()))
	tr.AddOutput(value.NewOutput(state.RequestChainAddress(), 1))
	// add state chain link
	tr.AddInput(value.NewInput(prevStateTx.Transfer().Id(), state.StateChainOutputIndex()))
	chainOutIdx := tr.AddOutput(value.NewOutput(state.StateChainAddress(), 1))

	nextState := sc.NewStateBlock(state.AssemblyId(), state.ConfigId(), reqTx.Id(), reqIdx)
	nextState.Builder().InitFromPrev(state)
	nextState.Builder().SetStateVars(nextStateVars)
	nextState.Builder().SetRequest(reqTx.Id(), reqIdx)
	nextState.Builder().SetStateChainOutputIndex(chainOutIdx)

	tx.SetState(nextState)
	return &resultCalculated{
		requestTx:    reqTx,
		requestIndex: reqIdx,
		state:        prevStateTx,
		resultTx:     tx,
	}, nil
}
