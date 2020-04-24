// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
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
	log.Infof("SC transaction received: %s", tx.Id().String())
}
