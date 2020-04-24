package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
)

// object with interface to build SC transaction and value transaction within it
// object panics if attempted to modify structure after finalization
type TransactionBuilder struct {
	inputs        *valuetransaction.Inputs
	outputs       map[address.Address][]*balance.Balance
	stateBlock    *StateBlock
	requestBlocks []*RequestBlock
	finalized     bool
}

func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		inputs:        valuetransaction.NewInputs(),
		outputs:       make(map[address.Address][]*balance.Balance),
		requestBlocks: make([]*RequestBlock, 0),
	}
}

func (txb *TransactionBuilder) Finalize() (*Transaction, error) {
	if txb.finalized {
		panic("attempt to modify already finalized transaction builder")
	}
	txv := valuetransaction.New(txb.inputs, valuetransaction.NewOutputs(txb.outputs))
	ret, err := NewTransaction(txv, txb.stateBlock, txb.requestBlocks)
	if err != nil {
		return nil, err
	}
	txb.finalized = true
	return ret, nil
}

func (txb *TransactionBuilder) AddInput(oid valuetransaction.OutputId) {
	txb.inputs.Add(oid)
}

func (txb *TransactionBuilder) AddOutput(addr address.Address, bal *balance.Balance) {
	if _, ok := txb.outputs[addr]; ok {
		txb.outputs[addr] = make([]*balance.Balance, 0)
	}
	txb.outputs[addr] = append(txb.outputs[addr], bal)
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
