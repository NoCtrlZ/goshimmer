package txdb

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"sync"
)

type localValueTxDb struct {
	sync.RWMutex
	byTxId                   map[HashValue]value.Transaction
	byTransferId             map[HashValue]value.Transaction
	spendingTxsByOutputRefId map[HashValue]value.Transaction
	outputsByAddress         map[HashValue][]*generic.OutputRef
}

const genesisAmount = 61 * parameters.Ti

func NewLocalDb() *localValueTxDb {
	ret := &localValueTxDb{
		byTxId:                   make(map[HashValue]value.Transaction),
		byTransferId:             make(map[HashValue]value.Transaction),
		spendingTxsByOutputRefId: make(map[HashValue]value.Transaction),
		outputsByAddress:         make(map[HashValue][]*generic.OutputRef),
	}
	genesisTransfer := value.NewUTXOTransfer()
	genesisTransfer.AddInput(value.NewInput(NilHash, 0))
	genesisTransfer.AddOutput(value.NewOutput(NilHash, genesisAmount))
	genesisTx := value.NewTransaction(genesisTransfer, nil)
	ret.byTxId[*NilHash] = genesisTx
	ret.byTransferId[*NilHash] = genesisTx
	ret.outputsByAddress[*NilHash] = []*generic.OutputRef{generic.NewOutputRef(NilHash, 0)}
	return ret
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
	err := ldb.__putTransaction(tx)
	ldb.Unlock()
	if err != nil {
		return err
	}
	fmt.Printf("++++ ldb: inserted new tx: id = %s trid = %s ShortStr: %s\n",
		tx.Id().Short(), tx.Transfer().Id().Short(), tx.Transfer().ShortStr())
	return nil
}

func (ldb *localValueTxDb) __putTransaction(tx value.Transaction) error {
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
	ldb.byTransferId[*trid] = tx
	ldb.byTxId[*tx.Id()] = tx
	// register each input as spent
	for _, inp := range tx.Transfer().Inputs() {
		ldb.spendingTxsByOutputRefId[*inp.OutputRef().Id()] = tx
	}
	// store outputs
	for i, outp := range tx.Transfer().Outputs() {
		addr := outp.Address()
		if _, ok := ldb.outputsByAddress[*addr]; !ok {
			ldb.outputsByAddress[*addr] = make([]*generic.OutputRef, 0)
		}
		ldb.outputsByAddress[*addr] = append(ldb.outputsByAddress[*addr], generic.NewOutputRef(trid, uint16(i)))
	}
	return nil
}

func (ldb *localValueTxDb) GetUnspentOutputs(addr *HashValue) []*generic.OutputRefWithAddrValue {
	outs, ok := ldb.outputsByAddress[*addr]
	if !ok {
		return nil
	}
	ret := make([]*generic.OutputRefWithAddrValue, 0)
	for _, out := range outs {
		if _, ok = ldb.spendingTxsByOutputRefId[*out.Id()]; !ok {
			ret = append(ret, value.MustGetOutputAddrValue(out))
		}
	}
	return ret
}
