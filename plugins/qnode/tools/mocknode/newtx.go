package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
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

func newOrigin(ownerAddr *hashing.HashValue) (sc.Transaction, error) {
	// create owners's transfer balance of 1i by transferring it from genesis
	keyPool := clientapi.NewDummyKeyPool()
	ret, err := clientapi.NewScOriginTransaction(clientapi.NewOriginParams{
		AssemblyId:      aid,
		ConfigId:        configId,
		AssemblyAccount: assemblyAccount,
		OwnerAccount:    ownerAddr,
	})
	err = sc.SignTransaction(ret, keyPool)

	if err != nil {
		return nil, err
	}
	err = sc.VerifySignatures(ret, keyPool)
	if err != nil {
		return nil, err
	}
	vtx, err := ret.ValueTx()
	if err != nil {
		return nil, err
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		return nil, err
	}
	return ret, err
}

func makeBetRequestTx(fromAccount *hashing.HashValue, betSum uint64, color int, reward uint64) (sc.Transaction, error) {

	fmt.Printf("+++ makeBetRequestTx: balance of %s is %d\n", fromAccount.Short(), value.GetBalance(fromAccount))

	vars := generic.NewFlatValueMap()
	vars.SetInt("req_type", fairroulette.REQ_TYPE_BET)
	vars.SetInt("color", color)

	ret, err := clientapi.NewRequestTransaction(clientapi.NewRequestParams{
		AssemblyId:       aid,
		AssemblyAccount:  assemblyAccount,
		RequesterAccount: fromAccount,
		Reward:           reward,
		Deposit:          betSum,
		Vars:             vars,
	})

	err = sc.SignTransaction(ret, keyPool)
	if err != nil {
		return nil, err
	}
	err = sc.VerifySignatures(ret, keyPool)
	if err != nil {
		panic(err)
	}
	fmt.Printf("+++ bet tansaction created for sum %d account %s\n", betSum, fromAccount.Short())
	return ret, nil
}
