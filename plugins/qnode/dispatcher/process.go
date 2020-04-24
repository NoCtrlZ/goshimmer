// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
)

// to be called for event attachment closure
func processIncomingValueTransaction(vtx *valuetransaction.Transaction) {
	tx, isScTransaction, err := sctransaction.ParseValueTransaction(vtx)
	if !isScTransaction {
		return //ignore
	}
	if err != nil {
		log.Errorf("error while parsing smart contract transaction %s: %v", vtx.Id().String(), err)
		return
	}
	log.Debugw("SC transaction received", "id", tx.Id().String())
	processState(tx)
	processRequests(tx)
}

func processState(tx *sctransaction.Transaction) {
	hasState, isOrigin, err := checkState(tx)
	if err != nil {
		log.Error(err)
		return
	}
	if !hasState {
		return
	}
	// all state block validations passed
	isOrigin = isOrigin
	// TODO dispatch state message
}

// validates and returns if it hash state, is it origin state or error
func checkState(tx *sctransaction.Transaction) (bool, bool, error) {
	stateBlock, stateExists := tx.State()
	if !stateExists {
		return false, false, nil
	}
	scid := stateBlock.ScId()
	balances, ok := tx.OutputBalancesByAddress(scid.Address())
	if !ok || len(balances) == 0 {
		// expected output of SC token to the SC address
		// invalid SC transaction. Ignore
		return false, false, fmt.Errorf("didn't find output to the SC address. tx id %s", tx.Id().String())
	}
	isOriginTx := false
	outBalance := sctransaction.SumBalancesOfColor(balances, scid.Color())
	if outBalance == 0 {
		// for origin transaction check COLOR_NEW
		outBalance = sctransaction.SumBalancesOfColor(balances, balance.COLOR_NEW)
		isOriginTx = true
	}
	if outBalance != 1 {
		return false, false, fmt.Errorf("non-existent or wrong output with SC token in sc tx %s", tx.Id().String())
	}
	if isOriginTx {
		if balance.Color(tx.Id()) != scid.Color() {
			return false, false, fmt.Errorf("for an origin sc transaction tx hash must be equal to the scid color. Inconsistent tx %s", tx.Id().String())
		}
	}
	return true, isOriginTx, nil
}

func processRequests(tx *sctransaction.Transaction) {
	// TODO
}
