package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

// constructors
var (
	newUTXOTransfer func() UTXOTransfer
	newInput        func(*hashing.HashValue, uint16) Input
	newOutput       func(*hashing.HashValue, uint64) Output
	newTx           func(UTXOTransfer, []byte) Transaction
	parseTx         func([]byte) (Transaction, error)
)

type SetConstructorsParams struct {
	UTXOConstructor   func() UTXOTransfer
	InputConstructor  func(*hashing.HashValue, uint16) Input
	OutputConstructor func(*hashing.HashValue, uint64) Output
	TxConstructor     func(UTXOTransfer, []byte) Transaction
	ParseConstructor  func([]byte) (Transaction, error)
}

func SetConstructors(par SetConstructorsParams) {
	newUTXOTransfer = par.UTXOConstructor
	newInput = par.InputConstructor
	newOutput = par.OutputConstructor
	newTx = par.TxConstructor
	parseTx = par.ParseConstructor
}

func NewUTXOTransfer() UTXOTransfer {
	return newUTXOTransfer()
}

func NewInput(transferId *hashing.HashValue, outputIndex uint16) Input {
	return newInput(transferId, outputIndex)
}

func NewOutput(address *hashing.HashValue, value uint64) Output {
	return newOutput(address, value)
}

func NewTransaction(transfer UTXOTransfer, payload []byte) Transaction {
	return newTx(transfer, payload)
}

func ParseTransaction(data []byte) (Transaction, error) {
	return parseTx(data)
}

type DB interface {
	GetValueTx(id *hashing.HashValue) Transaction
	GetTransfer(id *hashing.HashValue) UTXOTransfer
}

var valuetxdb DB

func SetValuetxDB(db DB) {
	valuetxdb = db
}

func GetValueTx(id *hashing.HashValue) Transaction {
	return valuetxdb.GetValueTx(id)
}

func GetTransfer(id *hashing.HashValue) UTXOTransfer {
	if id.Equal(hashing.NilHash) {
		return nil
	}
	return valuetxdb.GetTransfer(id)
}
