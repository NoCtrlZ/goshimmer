package txdb

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"sync"
)

type localValueTxDb struct {
	sync.RWMutex
	byTxId                   map[HashValue]value.Transaction
	byTransferId             map[HashValue]value.Transaction
	spendingTxsByOutputRefId map[HashValue]value.Transaction
	outputsByAddress         map[HashValue][]*generic.OutputRefWithValue
}

func NewLocalDb() *localValueTxDb {
	return &localValueTxDb{
		byTxId:                   make(map[HashValue]value.Transaction),
		byTransferId:             make(map[HashValue]value.Transaction),
		spendingTxsByOutputRefId: make(map[HashValue]value.Transaction),
		outputsByAddress:         make(map[HashValue][]*generic.OutputRefWithValue),
	}
}

func (ldb *localValueTxDb) GetByTransactionId(id *HashValue) (value.Transaction, bool) {
	ldb.RLock()
	defer ldb.RUnlock()

	ret, ok := ldb.byTxId[*id]
	return ret, ok
}

func (ldb *localValueTxDb) GetByTransferId(id *HashValue) (value.Transaction, bool) {
	ldb.RLock()
	defer ldb.RUnlock()

	ret, ok := ldb.byTransferId[*id]
	return ret, ok
}

func (ldb *localValueTxDb) GetSpendingTransaction(outputRefId *HashValue) (value.Transaction, bool) {
	ldb.RLock()
	defer ldb.RUnlock()

	ret, ok := ldb.spendingTxsByOutputRefId[*outputRefId]
	return ret, ok
}

func (ldb *localValueTxDb) PutTransaction(tx value.Transaction) error {
	ldb.Lock()
	defer ldb.Unlock()

	// check conflicts
	_, ok := ldb.byTxId[*tx.Id()]
	if ok {
		return fmt.Errorf("++++ conflict: another tx with id %s", tx.Id().Short())
	}
	trid := tx.Transfer().Id()
	_, ok = ldb.byTransferId[*trid]
	if ok {
		return fmt.Errorf("++++ conflict: another tx with transfer id %s", trid.Short())
	}
	for _, inp := range tx.Transfer().Inputs() {
		spendingTx, ok := ldb.spendingTxsByOutputRefId[*inp.OutputRef().Id()]
		if ok {
			return fmt.Errorf("++++ conflict: doublespend in tx id %s. Conflicts with tx %s",
				tx.Id().Short(), spendingTx.Id())
		}
	}
	ldb.byTxId[*tx.Id()] = tx
	ldb.byTransferId[*trid] = tx
	// register each input as spent
	for _, inp := range tx.Transfer().Inputs() {
		ldb.spendingTxsByOutputRefId[*inp.OutputRef().Id()] = tx
	}
	// store outputs
	for i, outp := range tx.Transfer().Outputs() {
		addr := outp.Address()
		if _, ok := ldb.outputsByAddress[*addr]; !ok {
			ldb.outputsByAddress[*addr] = make([]*generic.OutputRefWithValue, 0)
		}
		ldb.outputsByAddress[*addr] = append(ldb.outputsByAddress[*addr], &generic.OutputRefWithValue{
			OutputRef: *generic.NewOutputRef(trid, uint16(i)),
			Value:     outp.Value(),
		})
	}
	fmt.Printf("++++ ldb: inserted new tx: id = %s trid = %s\n", tx.Id().Short(), trid.Short())
	return nil
}

func (ldb *localValueTxDb) GetUnspentOutputs(addr *HashValue) []*generic.OutputRefWithValue {
	outs, ok := ldb.outputsByAddress[*addr]
	if !ok {
		return nil
	}
	ret := make([]*generic.OutputRefWithValue, 0)
	for _, outv := range outs {
		if _, ok = ldb.spendingTxsByOutputRefId[*outv.Id()]; !ok {
			ret = append(ret, outv)
		}
	}
	return ret
}
