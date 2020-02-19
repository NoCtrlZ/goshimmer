package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"math/rand"
	"net/http"
)

const webport = 2000

func runWebServer() {
	tools.Logf(0, "Web server is running on port %d", webport)
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
	aidString  = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	configName = "configuration 4"
)

var aid, configId, stateChainAddr, requestChainAddr, ownerChainAddr *hashing.HashValue

func init() {
	configId = hashing.HashStrings(configName)
	aid, _ = hashing.HashValueFromString(aidString)
	stateChainAddr = hashing.HashStrings("state_chain_addr")
	requestChainAddr = hashing.HashStrings("request_chain_addr")
	ownerChainAddr = hashing.HashStrings("owner_chain_addr")
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
