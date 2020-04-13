package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/modelimpl"
	"github.com/iotaledger/goshimmer/plugins/qnode/signedblock"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/txdb"
)

var (
	ownerAddress    *hashing.HashValue
	ownerTxPosted   bool
	ownerTx         sc.Transaction
	keyPool         generic.KeyPool
	ldb             value.DB
	assemblyAccount *hashing.HashValue
)

func initGlobals() {
	modelimpl.Init()
	signedblock.Init()

	assemblyAccount = params.Addresses[0]

	ldb = txdb.NewLocalDb(nil)
	value.SetValuetxDB(ldb)

	keyPool = clientapi.NewDummyKeyPool()
	// owner account with 1i
	ownerAddress = hashing.HashData(params.Scid.Bytes())
	fmt.Printf("owner's account will be %s\n", ownerAddress.Short())
	ownerTx, _ = generateAccountWithDeposit(ownerAddress, 1)
}

func generateAccountWithDeposit(addr *hashing.HashValue, deposit uint64) (sc.Transaction, *generic.OutputRef) {
	tx := sc.NewTransaction()

	// taking funds from genesis
	outIdx, err := clientapi.MoveFundsFromToAddress(tx, hashing.NilHash, addr, []uint64{deposit})
	if err != nil {
		panic(err)
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		panic(err)
	}
	err = sc.SignTransaction(tx, keyPool)
	if err != nil {
		panic(err)
	}
	err = sc.VerifySignatures(tx, keyPool)
	if err != nil {
		panic(err)
	}
	err = ldb.PutTransaction(vtx)
	if err != nil {
		panic(err)
	}
	return tx, generic.NewOutputRef(tx.Transfer().Id(), outIdx[0])
}
