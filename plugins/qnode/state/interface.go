package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

type VariableState interface {
	StateIndex() uint32
	Apply(StateUpdate) VariableState
	Bytes() []byte
	SaveToDb() error
}

type StateUpdate interface {
	StateIndex() uint32
	StateTransactionId() valuetransaction.Id
	SetStateTransactionId(valuetransaction.Id)
	Bytes() []byte
	SaveToDb() error
}
