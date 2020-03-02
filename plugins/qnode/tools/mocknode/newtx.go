package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"math/rand"
)

const (
	assemblyDescription = "test assembly 1"
	N                   = uint16(4)
	T                   = uint16(3)
	cfgId2              = "74ca4be8414a6d787dd54d2452364b59c88c11b23a768b8d12810a057fb9a777" // 2 accounts
	cfgId10             = "d129cdd27f69df6d12f6bf8d4df377966ea410d22484c4addf3dcbc7080de30c" // 2 accounts
)

var accStrings2 = []string{
	"c59de480c9ea21705b0d66299f14e9976308e3d7802971271b5eedd9e1f7a9ad",
	"158284bb4c1f33342681832bed2b807286744f098f7f1c58289169ba7b603415",
}

var accStringsN10 = []string{
	"ceb5579e21e651dd48c47eea42fa7e6ddd0732e3df9ef8de127d693b977ea4e1",
	"60ef310872f2b4d09cb2fa43e843b514fc21d3ea72b268a39d822b8ca9d5fd19",
}

var aid, configId, assemblyAccount, ownerAddr *hashing.HashValue

func init() {
	configId, _ = hashing.HashValueFromString(cfgId2)
	aid = hashing.HashStrings(assemblyDescription)
	assemblyAccount, _ = hashing.HashValueFromString(accStrings2[0])
	ownerAddr = hashing.NilHash
}

func newOrigin() (sc.Transaction, error) {
	// create owners's transfer of 1i to owner's address
	toorig := sc.NewTransaction()
	trf := toorig.Transfer()
	trf.AddInput(value.NewInput(hashing.RandomHash(nil), 0))
	outIdx := trf.AddOutput(value.NewOutput(ownerAddr, 1))

	keyPool := clientapi.NewDummyKeyPool()
	err := sc.SignTransaction(toorig, keyPool)
	if err != nil {
		return nil, err
	}
	err = sc.VerifySignedBlocks(toorig.Signatures(), keyPool)
	if err != nil {
		panic(err)
	}

	vtx, err := toorig.ValueTx()
	if err != nil {
		return nil, err
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		return nil, err
	}
	origOutRef := generic.NewOutputRef(toorig.Transfer().Id(), outIdx)
	ret, err := clientapi.NewOriginTransaction(clientapi.NewOriginParams{
		AssemblyId:      aid,
		ConfigId:        configId,
		AssemblyAccount: assemblyAccount,
		OwnerAccount:    ownerAddr,
		OriginOutput:    origOutRef,
	})

	err = sc.SignTransaction(ret, keyPool)

	if err != nil {
		return nil, err
	}
	err = sc.VerifySignedBlocks(ret.Signatures(), keyPool)
	if err != nil {
		panic(err)
	}
	return ret, err
}

var reqnrseq = 0

func makeReqTx() (sc.Transaction, error) {
	vars := map[string]string{
		"reqnr": fmt.Sprintf("#%d", reqnrseq),
		"salt":  fmt.Sprintf("%d", rand.Int()),
	}
	reqnrseq++
	reqestorAccount := hashing.RandomHash(nil)

	toreq := sc.NewTransaction()
	toreq.Transfer().AddInput(value.NewInput(reqestorAccount, 0))
	outIdx := toreq.Transfer().AddOutput(value.NewOutput(reqestorAccount, 1))

	keyPool := clientapi.NewDummyKeyPool()
	err := sc.SignTransaction(toreq, keyPool)
	if err != nil {
		return nil, err
	}
	err = sc.VerifySignedBlocks(toreq.Signatures(), keyPool)
	if err != nil {
		panic(err)
	}
	vtx, err := toreq.ValueTx()
	if err != nil {
		return nil, err
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		return nil, err
	}
	reqChainOut := generic.NewOutputRef(toreq.Transfer().Id(), outIdx)

	ret, err := clientapi.NewRequest(clientapi.NewRequestParams{
		AssemblyId:         aid,
		AssemblyAccount:    assemblyAccount,
		RequestChainOutput: reqChainOut,
		Vars:               vars,
	})

	err = sc.SignTransaction(ret, keyPool)
	if err != nil {
		return nil, err
	}
	err = sc.VerifySignedBlocks(ret.Signatures(), keyPool)
	if err != nil {
		panic(err)
	}

	return ret, err
}

var players []*hashing.HashValue

const numPlayers = 5

var curPlayer = 0

func init() {
	players = make([]*hashing.HashValue, numPlayers)
	for i := range players {
		players[i] = hashing.RandomHash(nil)
	}
}

//
//func makeBetRequestTx(bet uint64) (sc.Transaction, error){
//	playersAddr := players[curPlayer]
//	curPlayer++
//	betSum := uint64(1000000)
//	reward := uint64(2000)
//
//	// create requestor's output for the sum
//	toreq := sc.NewTransaction()
//	toreq.Transfer().AddInput(value.NewInput(hashing.NilHash, 0))
//	outIdx := toreq.Transfer().AddOutput(value.NewOutput(playersAddr, betSum+reward+1))
//
//	keyPool := clientapi.NewDummyKeyPool()
//	err := sc.SignTransaction(toreq, keyPool)
//	if err != nil {
//		return nil, err
//	}
//	err = sc.VerifySignedBlocks(toreq.Signatures(), keyPool)
//	if err != nil {
//		panic(err)
//	}
//	vtx, err := toreq.ValueTx()
//	if err != nil {
//		return nil, err
//	}
//	if err := ldb.PutTransaction(vtx); err != nil {
//		return nil, err
//	}
//	reqChainOut := generic.NewOutputRef(toreq.Transfer().Id(), outIdx)
//
//	ret, err := clientapi.NewRequest(clientapi.NewRequestParams{
//		AssemblyId:         aid,
//		AssemblyAccount:    assemblyAccount,
//		RequestChainOutput: reqChainOut,
//		Vars:               vars,
//	})
//
//	err = sc.SignTransaction(ret, keyPool)
//	if err != nil {
//		return nil, err
//	}
//	err = sc.VerifySignedBlocks(ret.Signatures(), keyPool)
//	if err != nil {
//		panic(err)
//	}
//
//	return ret, err
//
//}
