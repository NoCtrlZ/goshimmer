package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"io"
)

type VariableState interface {
	StateIndex() uint32
	Apply(StateUpdate) VariableState
	SaveToDb() error
	Read(io.Reader) error
	Write(io.Writer) error
}

// state update with anchor transaction hash
type StateUpdate interface {
	ScId() sctransaction.ScId
	StateIndex() uint32
	StateTransactionId() valuetransaction.Id
	SetStateTransactionId(valuetransaction.Id)
	SaveToDb() error
	Read(io.Reader) error
	Write(io.Writer) error
}
