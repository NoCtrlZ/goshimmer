package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"io"
)

type VariableState interface {
	StateIndex() uint32
	Apply(StateUpdate) VariableState
	SaveToDb() error
	Read(io.Reader) error
	Write(io.Writer) error
}

type StateUpdate interface {
	StateIndex() uint32
	StateTransactionId() valuetransaction.Id
	SetStateTransactionId(valuetransaction.Id)
	SaveToDb() error
	Read(io.Reader) error
	Write(io.Writer) error
}
