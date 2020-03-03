package main

import (
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
	requesterAddresses []*hashing.HashValue
	requesterDeposits  []*generic.OutputRef
	keyPool            generic.KeyPool
	reqnrseq           = 0
	curPlayer          = 0
	ldb                value.DB
)

const (
	numPlayers  = 5
	initDeposit = uint64(10000000)
)

func initGlobals() {
	modelimpl.Init()
	signedblock.Init()

	ldb = txdb.NewLocalDb()
	value.SetValuetxDB(ldb)

	keyPool = clientapi.NewDummyKeyPool()
	requesterAddresses = make([]*hashing.HashValue, numPlayers)
	requesterDeposits = make([]*generic.OutputRef, numPlayers)
	for i := range requesterAddresses {
		requesterAddresses[i] = hashing.RandomHash(nil)
	}
	// deposit fake iotas to addresses owned by requesterAddresses
	keyPool = clientapi.NewDummyKeyPool()
	for i, addr := range requesterAddresses {
		requesterDeposits[i] = generateAccountWithDeposit(addr, initDeposit)
	}
}

func generateAccountWithDeposit(addr *hashing.HashValue, deposit uint64) *generic.OutputRef {
	tx := sc.NewTransaction()
	tx.Transfer().AddInput(value.NewInput(hashing.RandomHash(nil), 0))
	outIdx := tx.Transfer().AddOutput(value.NewOutput(addr, deposit))
	vtx, err := tx.ValueTx()
	if err != nil {
		panic(err)
	}
	err = sc.SignTransaction(tx, keyPool)
	if err != nil {
		panic(err)
	}
	err = sc.VerifySignedBlocks(tx.Signatures(), keyPool)
	if err != nil {
		panic(err)
	}
	err = ldb.PutTransaction(vtx)
	if err != nil {
		panic(err)
	}
	return generic.NewOutputRef(tx.Transfer().Id(), outIdx)
}
