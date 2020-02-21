package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
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

var aid, configId, stateAddr, requestAddr, ownerAddr *hashing.HashValue

func init() {
	configId, _ = hashing.HashValueFromString(cfgId2)
	aid = hashing.HashStrings(assemblyDescription)
	stateAddr, _ = hashing.HashValueFromString(accStrings2[0])
	requestAddr, _ = hashing.HashValueFromString(accStrings2[0])
	ownerAddr = hashing.NilHash
}

func newOrigin() (sc.Transaction, error) {
	return clientapi.NewOriginTransaction(clientapi.NewOriginParams{
		AssemblyId:     aid,
		ConfigId:       configId,
		StateAccount:   stateAddr,
		RequestAccount: requestAddr,
		OwnerAccount:   ownerAddr,
		OriginOutput:   generic.NewOutputRef(hashing.NilHash, 0),
	})
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
	return clientapi.NewRequest(clientapi.NewRequestParams{
		AssemblyId:         aid,
		RequestAccount:     requestAddr,
		RequestChainOutput: generic.NewOutputRef(hashing.NilHash, 0),
		Vars:               vars,
	})
}
