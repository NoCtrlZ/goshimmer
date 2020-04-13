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

func newOrigin(ownerAddr *hashing.HashValue) (sc.Transaction, error) {
	// create owners's transfer balance of 1i by transferring it from genesis
	keyPool := clientapi.NewDummyKeyPool()
	ret, err := clientapi.NewScOriginTransaction(clientapi.NewOriginParams{
		AssemblyId:      params.Scid,
		ConfigId:        params.ConfigId,
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
		AssemblyId:       params.Scid,
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
