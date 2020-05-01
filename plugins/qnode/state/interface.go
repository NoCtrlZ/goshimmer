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

// state update without anchor transaction hash
type StateUpdateEssence interface {
	StateIndex() uint32
	Read(io.Reader) error
	Write(io.Writer) error
}

// state update with anchor transaction hash
type StateUpdate interface {
	Essence() StateUpdateEssence
	StateTransactionId() valuetransaction.Id
	SetStateTransactionId(valuetransaction.Id)
	IsAnchored() bool
	SaveToDb() error
	Read(io.Reader) error
	Write(io.Writer) error
}
