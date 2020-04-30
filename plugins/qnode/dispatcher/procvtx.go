// this package defines main entry how value transactions are entering the qnode
package dispatcher

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"hash/crc32"
)

func processIncomingValueTransaction(vtx *valuetransaction.Transaction) {
	// quick check to filter out those which are definitely not interesting
	if transactionToBeIgnored(vtx) {
		return
	}
	// transaction must be parsed
	tx, isScTransaction, err := sctransaction.ParseValueTransaction(vtx)
	if !isScTransaction {
		return //ignore
	}
	if err != nil {
		log.Errorf("error while parsing smart contract transaction %s: %v", vtx.Id().String(), err)
		return
	}
	log.Debugw("SC transaction received", "id", tx.Id().String())
	dispatchState(tx)
	dispatchRequests(tx)
}

// recognizes if the payload can be a parsed as SC payload and it is of interest
// without parsing the whole data payload
// returns true if it is to be ignored in following situations
//  - too short
//  - wrong checksum
//  - it doesn't have request blocks and scid is not processed by this node
func transactionToBeIgnored(vtx *valuetransaction.Transaction) bool {
	// 1 for meta byte
	// 4 for checksum
	// 65 for scid of ths first block (it may be state or request)
	// minimum sc data payload size is 70 bytes
	data := vtx.GetDataPayload()
	if len(data) < 1+4+sctransaction.ScIdLength {
		// too short for sc payload
		return true
	}
	checksumGiven := util.Uint32From4Bytes(data[1 : 1+4])
	checksumCalculated := crc32.ChecksumIEEE(data[1+4 : 1+4+sctransaction.ScIdLength])
	if checksumGiven != checksumCalculated {
		// wrong checksum, not a SC transaction
		return true
	}
	// check transaction which only have state if it is processed by this node
	hasState, numRequests := sctransaction.DecodeMetaByte(data[0])
	if hasState && numRequests == 0 {
		// check the color of state in the dictionary of SCs processed by the node
		col, _ := sctransaction.ColorFromBytes(data[1+4 : 1+4+balance.ColorLength])
		if getCommittee(col) == nil {
			// it may be a valid sc transaction, but definitely not interesting for this node
			return true
		}
	}
	// transaction must be parsed and processed
	return false
}

// validates and returns if it has state block, is it origin state or error
func validateState(tx *sctransaction.Transaction) (bool, error) {
	stateBlock, stateExists := tx.State()
	if !stateExists {
		return false, nil
	}
	scid := stateBlock.ScId()
	balances, ok := tx.OutputBalancesByAddress(scid.Address())
	if !ok || len(balances) == 0 {
		// expected output of SC token to the SC address
		// invalid SC transaction. Ignore
		return false, fmt.Errorf("didn't find output to the SC address. tx id %s", tx.Id().String())
	}
	isOriginTx := false
	outBalance := sctransaction.SumBalancesOfColor(balances, scid.Color())
	if outBalance == 0 {
		// for the origin transaction check COLOR_NEW
		outBalance = sctransaction.SumBalancesOfColor(balances, balance.ColorNew)
		isOriginTx = true
	}
	if outBalance != 1 {
		// supply of the SC token must be exactly 1
		return false, fmt.Errorf("non-existent or wrong output with SC token in sc tx %s", tx.Id().String())
	}
	if isOriginTx {
		// if this is an origin tx, the the hash (Id) of the transaction must be equal
		// to the color of the scid
		if balance.Color(tx.Id()) != scid.Color() {
			return false, fmt.Errorf("for an origin sc transaction tx hash must be equal to the scid color. Inconsistent tx %s", tx.Id().String())
		}
	}
	return true, nil
}

func dispatchState(tx *sctransaction.Transaction) {
	hasState, err := validateState(tx)
	if err != nil {
		log.Error(err)
		return
	}
	if !hasState {
		return
	}
	// all state block validations passed
	if committee := getCommittee(tx.MustState().ScId().Color()); committee != nil {
		committee.ProcessMessage(commtypes.StateTransactionMsg{Transaction: tx})
	}
}

func dispatchRequests(tx *sctransaction.Transaction) {
	for i, reqBlk := range tx.Requests() {
		if committee := getCommittee(reqBlk.ScId().Color()); committee != nil {
			committee.ProcessMessage(commtypes.RequestMsg{
				Transaction: tx,
				Index:       uint16(i),
			})
		}
	}
}
