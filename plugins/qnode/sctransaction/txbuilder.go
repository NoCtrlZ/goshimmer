package sctransaction

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

// object with interface to build SC transaction and value transaction within it
// object panics if attempted to modify structure after finalization
type TransactionBuilder struct {
	inputs        *valuetransaction.Inputs
	outputs       map[address.Address]map[balance.Color]int64
	stateBlock    *StateBlock
	requestBlocks []*RequestBlock
	finalized     bool
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		inputs:        valuetransaction.NewInputs(),
		outputs:       make(map[address.Address]map[balance.Color]int64),
		requestBlocks: make([]*RequestBlock, 0),
	}
}

func (txb *TransactionBuilder) Finalize() (*Transaction, error) {
	if txb.finalized {
		return nil, errors.New("attempt to modify already finalized transaction builder")
	}

	outputs := make(map[address.Address][]*balance.Balance)
	for addr, bmap := range txb.outputs {
		outputs[addr] = make([]*balance.Balance, 0)
		for col, val := range bmap {
			outputs[addr] = append(outputs[addr], balance.New(col, val))
		}
	}
	txv := valuetransaction.New(txb.inputs, valuetransaction.NewOutputs(outputs))
	ret, err := NewTransaction(txv, txb.stateBlock, txb.requestBlocks)
	if err != nil {
		return nil, err
	}
	txb.finalized = true
	return ret, nil
}

func (txb *TransactionBuilder) AddInputs(oid ...valuetransaction.OutputId) {
	for _, o := range oid {
		txb.inputs.Add(o)
	}
}

func (txb *TransactionBuilder) AddBalanceToOutput(addr address.Address, bal *balance.Balance) {
	if _, ok := txb.outputs[addr]; ok {
		txb.outputs[addr] = make(map[balance.Color]int64)
	}
	balances := txb.outputs[addr]
	if val, ok := balances[bal.Color()]; ok {
		balances[bal.Color()] = val + bal.Value()
	} else {
		balances[bal.Color()] = bal.Value()
	}
}

func (txb *TransactionBuilder) AddStateBlock(scid *ScId, stateIndex uint32) {
	txb.stateBlock = NewStateBlock(scid, stateIndex)
}

func (txb *TransactionBuilder) SetStateBlockParams(params StateBlockParams) {
	txb.stateBlock.WithParams(params)
}

func (txb *TransactionBuilder) AddRequestBlock(reqBlk *RequestBlock) {
	txb.requestBlocks = append(txb.requestBlocks, reqBlk)
}
