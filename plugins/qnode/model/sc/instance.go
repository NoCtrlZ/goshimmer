package sc

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
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
