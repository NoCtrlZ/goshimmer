package main

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

type localValueTxDb struct {
	byTxId                   map[HashValue]value.Transaction
	byTransferId             map[HashValue]value.Transaction
	spendingTxsByOutputRefId map[HashValue]value.Transaction
}

func newLocalDb() *localValueTxDb {
	return &localValueTxDb{
		byTxId:                   make(map[HashValue]value.Transaction),
		byTransferId:             make(map[HashValue]value.Transaction),
		spendingTxsByOutputRefId: make(map[HashValue]value.Transaction),
	}
}

func (ldb *localValueTxDb) GetByTransactionId(id *HashValue) (value.Transaction, bool) {
	ret, ok := ldb.byTxId[*id]
	return ret, ok
}

func (ldb *localValueTxDb) GetByTransferId(id *HashValue) (value.Transaction, bool) {
	ret, ok := ldb.byTransferId[*id]
	return ret, ok
}

func (ldb *localValueTxDb) GetSpendingTransaction(outputRefId *HashValue) (value.Transaction, bool) {
	ret, ok := ldb.spendingTxsByOutputRefId[*outputRefId]
	return ret, ok
}

func (ldb *localValueTxDb) PutTransaction(tx value.Transaction) bool {
	// check conflicts
	_, ok := ldb.GetByTransactionId(tx.Id())
	if ok {
		return false
	}
	trid := tx.Transfer().Id()
	_, ok = ldb.GetByTransferId(trid)
	if ok {
		return false
	}
	for _, inp := range tx.Transfer().Inputs() {
		_, ok = ldb.spendingTxsByOutputRefId[*inp.OutputRef().Id()]
		if ok {
			return false // corresponding output already spent, tx is conflicting
		}
	}
	ldb.byTxId[*tx.Id()] = tx
	ldb.byTransferId[*trid] = tx
	// register each input as spent
	for _, inp := range tx.Transfer().Inputs() {
		ldb.spendingTxsByOutputRefId[*inp.OutputRef().Id()] = tx
	}
	return true
}
