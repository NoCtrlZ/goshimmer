package sc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

// constructors

var (
	newTransaction  func() Transaction
	newFromValueTx  func(value.Transaction) (Transaction, error)
	newStateBlock   func(*hashing.HashValue, *hashing.HashValue, *hashing.HashValue, uint16) State
	newRequestBlock func(*hashing.HashValue, bool) Request
)

type SetConstructorsParams struct {
	TxConstructor           func() Transaction
	TxParser                func(value.Transaction) (Transaction, error)
	StateBlockConstructor   func(*hashing.HashValue, *hashing.HashValue, *hashing.HashValue, uint16) State
	RequestBlockConstructor func(*hashing.HashValue, bool) Request
}

func SetConstructors(c SetConstructorsParams) {
	newTransaction = c.TxConstructor
	newFromValueTx = c.TxParser
	newStateBlock = c.StateBlockConstructor
	newRequestBlock = c.RequestBlockConstructor
}

func NewTransaction() Transaction {
	return newTransaction()
}

func ParseTransaction(vtx value.Transaction) (Transaction, error) {
	return newFromValueTx(vtx)
}

func NewStateBlock(aid, cid *hashing.HashValue, reqRef *RequestRef) State {
	if reqRef == nil {
		return newStateBlock(aid, cid, hashing.NilHash, 0)
	}
	return newStateBlock(aid, cid, reqRef.Tx().Id(), reqRef.Index())
}

func NewRequestBlock(aid *hashing.HashValue, isConfig bool) Request {
	return newRequestBlock(aid, isConfig)
}

func NextStateUpdateTransaction(stateTx Transaction, reqRef *RequestRef) (Transaction, error) {
	state, ok := stateTx.State()
	if !ok {
		return nil, fmt.Errorf("NextStateUpdateTransaction: state block not found")
	}
	reqBlock := reqRef.RequestBlock()
	// check if request block points to valid chain output
	// which can be used as request->result chain
	requestChainOutputIdx, _, _ := reqBlock.OutputIndices()
	requestChainOutputRef := generic.NewOutputRef(reqRef.Tx().Transfer().Id(), requestChainOutputIdx)
	if !value.OutputCanBeChained(requestChainOutputRef, state.Config().AssemblyAccount()) {
		return nil, fmt.Errorf("invalid request chain output")
	}
	tx := NewTransaction()
	tr := tx.Transfer()
	// add request chain link
	// transfer 1i from RequestChainAddress to itself
	tr.AddInput(value.NewInputFromOutputRef(requestChainOutputRef))
	tr.AddOutput(value.NewOutput(state.Config().AssemblyAccount(), 1))

	// add state chain link
	// transfer 1i from StateChainAddress to itself
	tr.AddInput(value.NewInput(stateTx.Transfer().Id(), state.StateChainOutputIndex()))
	chainOutIdx := tr.AddOutput(value.NewOutput(state.Config().AssemblyAccount(), 1))

	nextState := NewStateBlock(state.AssemblyId(), state.Config().Id(), reqRef)
	nextState.
		WithStateIndex(state.StateIndex() + 1).
		WithVars(state.Vars()).
		WithStateChainOutputIndex(chainOutIdx)
	nextState.Config().With(state.Config())
	tx.SetState(nextState)
	return tx, nil
}
