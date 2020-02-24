package clientapi

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

type NewOriginParams struct {
	AssemblyId     *HashValue
	ConfigId       *HashValue
	StateAccount   *HashValue
	RequestAccount *HashValue
	OwnerAccount   *HashValue
	// owner's section
	OriginOutput *generic.OutputRef // output of 1i to the owner's address
}

// transfer is not signed

func NewOriginTransaction(par NewOriginParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	state := sc.NewStateBlock(par.AssemblyId, par.ConfigId, nil)
	configVars := state.ConfigVars()
	configVars.SetString(sc.MAP_KEY_STATE_ACCOUNT, par.StateAccount.String())
	configVars.SetString(sc.MAP_KEY_REQUEST_ACCOUNT, par.RequestAccount.String())
	configVars.SetString(sc.MAP_KEY_OWNER_ACCOUNT, par.OwnerAccount.String())
	ret.SetState(state)
	tr := ret.Transfer()

	// adding owner chain: transfer of 1i from owner's account to the stateAccount
	// the latter will be used to build chain
	addr, val := value.GetAddrValue(par.OriginOutput)
	if !addr.Equal(par.OwnerAccount) || val != 1 {
		return nil, fmt.Errorf("OriginOutput parameter must be exactly 1i to the owner's account")
	}
	tr.AddInput(value.NewInputFromOutputRef(par.OriginOutput))
	tr.AddOutput(value.NewOutput(par.StateAccount, 1))
	return ret, nil
}

type NewRequestParams struct {
	AssemblyId         *HashValue
	RequestAccount     *HashValue
	RequestChainOutput *generic.OutputRef // output of 1i owned by the request originator
	Vars               map[string]string
}

func NewRequest(par NewRequestParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	tr := ret.Transfer()
	// create 1i transfer from RequestChainOutput to request account
	tr.AddInput(value.NewInputFromOutputRef(par.RequestChainOutput))
	chainOutIndex := tr.AddOutput(value.NewOutput(par.RequestAccount, 1))
	_, val := value.GetAddrValue(par.RequestChainOutput)
	if val != 1 {
		return nil, fmt.Errorf("request chain output must have value exactly 1i")
	}
	reqBlk := sc.NewRequestBlock(par.AssemblyId, false, chainOutIndex)
	vars := reqBlk.Vars()
	for k, v := range par.Vars {
		vars.SetString(generic.VarName(k), v)
	}
	ret.AddRequest(reqBlk)
	return ret, nil
}
