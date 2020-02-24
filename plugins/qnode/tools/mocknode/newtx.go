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
)

var accStrings2 = []string{
	"c59de480c9ea21705b0d66299f14e9976308e3d7802971271b5eedd9e1f7a9ad",
	"158284bb4c1f33342681832bed2b807286744f098f7f1c58289169ba7b603415",
}

var aid, configId, stateAddr, requestAddr, ownerAddr, requestorAddr *hashing.HashValue

func init() {
	configId, _ = hashing.HashValueFromString(cfgId2)
	aid = hashing.HashStrings(assemblyDescription)
	stateAddr, _ = hashing.HashValueFromString(accStrings2[0])
	requestAddr, _ = hashing.HashValueFromString(accStrings2[0])
	requestorAddr = hashing.RandomHash(nil)
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
		AssemblyId:     aid,
		ConfigId:       configId,
		StateAccount:   stateAddr,
		RequestAccount: requestAddr,
		OwnerAccount:   ownerAddr,
		OriginOutput:   origOutRef,
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

func makeReqTx(reqnr string) (sc.Transaction, error) {
	reqnrstr := fmt.Sprintf("#%d", reqnrseq)
	if reqnr == "_seq" {
		reqnrseq++
	} else {
		reqnrstr = reqnr
	}
	vars := map[string]string{
		"reqnr": reqnrstr,
		"salt":  fmt.Sprintf("%d", rand.Int()),
	}
	// create owners's transfer of 1i to owner's address
	toreq := sc.NewTransaction()
	trf := toreq.Transfer()
	trf.AddInput(value.NewInput(hashing.RandomHash(nil), 0))
	outIdx := trf.AddOutput(value.NewOutput(requestAddr, 1))

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
		RequestAccount:     requestAddr,
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
