package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

// interface with ValueTangle

type InputRef struct {
	Tx    Transaction
	Index uint16
}

type DB interface {
	PutTransaction(tx Transaction) error
	GetByTransactionId(id *hashing.HashValue) (Transaction, bool)
	GetByTransferId(id *hashing.HashValue) (Transaction, bool)
	GetSpendingTransaction(outputRefId *hashing.HashValue) (Transaction, bool)
	GetUnspentOutputs(addr *hashing.HashValue) []*generic.OutputRefWithAddrValue
}

var valuetxdb DB

func SetValuetxDB(db DB) {
	valuetxdb = db
}

func GetTransfer(id *hashing.HashValue) UTXOTransfer {
	tx, ok := valuetxdb.GetByTransferId(id)
	if !ok {
		return nil
	}
	return tx.Transfer()
}

func PutTransaction(tx Transaction) error {
	return valuetxdb.PutTransaction(tx)
}

func GetByTransactionId(id *hashing.HashValue) (Transaction, bool) {
	return valuetxdb.GetByTransactionId(id)
}

func GetByTransferId(id *hashing.HashValue) (Transaction, bool) {
	return valuetxdb.GetByTransferId(id)
}

func GetUnspentOutputs(addr *hashing.HashValue) []*generic.OutputRefWithAddrValue {
	return valuetxdb.GetUnspentOutputs(addr)
}
