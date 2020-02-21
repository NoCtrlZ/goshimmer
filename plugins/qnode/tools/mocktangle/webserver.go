package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"math/rand"
	"net/http"
)

const webport = 2000

func runWebServer() {
	fmt.Printf("Web server is running on port %d\n", webport)
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", webport), nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	var tx sc.Transaction
	isCfgStr := r.FormValue("cfg")
	if isCfgStr != "" {
		tx = makeTestConfiguration()
	} else {
		reqnr := r.FormValue("reqnr")
		if reqnr == "" {
			_, _ = fmt.Fprintf(w, "reqnr parameter not provided")
			return
		}
		_, _ = fmt.Fprintf(w, "Received reqnr = %s", reqnr)
		tx = makeReqTx(reqnr)
	}
	postMsg(tx)
}

const (
	aidString = "test assembly 1"
	cfgId1    = "d85036600bb75389dae0d501d983bbe0d1edb3251a5590816c314d9f390cb85f" // 1 account
	cfgId2    = "eddb2656a97ff6be411aac0d2fddb1fd1cc7de42905eaa742a09031ee921c261" // 2 accounts
)

var accStrings2 = []string{
	"c59de480c9ea21705b0d66299f14e9976308e3d7802971271b5eedd9e1f7a9ad",
	"158284bb4c1f33342681832bed2b807286744f098f7f1c58289169ba7b603415",
}

var aid, configId, stateChainAddr, requestChainAddr, ownerChainAddr *hashing.HashValue

func init() {
	configId, _ = hashing.HashValueFromString(cfgId2)
	aid, _ = hashing.HashValueFromString(aidString)
	stateChainAddr, _ = hashing.HashValueFromString(accStrings2[0])
	requestChainAddr, _ = hashing.HashValueFromString(accStrings2[0])
	ownerChainAddr, _ = hashing.HashValueFromString(accStrings2[1])
}

func makeTestConfiguration() sc.Transaction {
	ret := sc.NewTransaction()
	state := sc.NewStateBlock(aid, configId, hashing.NilHash, 0)
	configVars := state.ConfigVars()
	configVars.SetString("state_chain_addr", stateChainAddr.String())
	configVars.SetString("request_chain_addr", requestChainAddr.String())
	configVars.SetString("owner_chain_addr", ownerChainAddr.String())
	ret.SetState(state)
	tr := ret.Transfer()
	tr.AddInput(value.NewInput(hashing.NilHash, 0))
	tr.AddOutput(value.NewOutput(stateChainAddr, 1))
	sigs := tr.InputSignatures()
	sig, ok := sigs[*hashing.NilHash]
	if !ok {
		panic("too bad")
	}
	sig.SetSignature(hashing.NilHash.Bytes(), generic.SIG_TYPE_FAKE)

	return ret
}

var reqnrseq = 0

func makeReqTx(reqnr string) sc.Transaction {
	ret := sc.NewTransaction()
	tr := ret.Transfer()
	tr.AddInput(value.NewInput(hashing.NilHash, 0))
	tr.AddOutput(value.NewOutput(requestChainAddr, 1))
	sigs := tr.InputSignatures()

	sig, ok := sigs[*hashing.NilHash]
	if !ok {
		panic("too bad")
	}
	sig.SetSignature(hashing.NilHash.Bytes(), generic.SIG_TYPE_FAKE)

	reqBlk := sc.NewRequestBlock(aid, false)
	vars := reqBlk.Vars()

	// TODO add and sign transfer from my addr.

	if reqnr == "_seq" {
		vars.SetString("reqnr", fmt.Sprintf("#%d", reqnrseq))
		reqnrseq++
	} else {
		vars.SetString("reqnr", reqnr)
	}
	vars.SetInt("salt", rand.Int()) // random salt to make the request unique
	ret.AddRequest(reqBlk)

	return ret
}
