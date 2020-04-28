package state

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
)

type VariableState interface {
	Apply(StateUpdate) VariableState
	Bytes() []byte
	LoadFromDb(balance.Color) error
	SaveToDb(balance.Color) error
}

type StateUpdate interface {
	StateTransactionId() transaction.Id
	SetStateTransactionId(transaction.Id)
	Bytes() []byte
	LoadFromDb(balance.Color, uint32) error
	SaveToDb(balance.Color, uint32) error
}
