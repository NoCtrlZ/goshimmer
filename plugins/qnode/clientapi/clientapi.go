package clientapi

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

type NewOriginParams struct {
	AssemblyId      *HashValue
	ConfigId        *HashValue
	AssemblyAccount *HashValue
	OwnerAccount    *HashValue
	// owner's section
	OriginOutput *generic.OutputRef // output of 1i to the owner's address
}

// transfer is not signed

func NewOriginTransaction(par NewOriginParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	state := sc.NewStateBlock(par.AssemblyId, par.ConfigId, nil)
	configVars := state.Config().Vars()
	configVars.SetString(sc.MAP_KEY_ASSEMBLY_ACCOUNT, par.AssemblyAccount.String())
	configVars.SetString(sc.MAP_KEY_OWNER_ACCOUNT, par.OwnerAccount.String())
	ret.SetState(state)
	tr := ret.Transfer()

	// adding owner chain: transfer of 1i from owner's account to the stateAccount
	// the latter will be used to build chain
	addr, val := value.MustGetOutputAddrValue(par.OriginOutput)
	if !addr.Equal(par.OwnerAccount) || val != 1 {
		return nil, fmt.Errorf("OriginOutput parameter must be exactly 1i to the owner's account")
	}
	tr.AddInput(value.NewInputFromOutputRef(par.OriginOutput))
	tr.AddOutput(value.NewOutput(par.AssemblyAccount, 1))
	return ret, nil
}

type NewRequestParams struct {
	AssemblyId         *HashValue
	AssemblyAccount    *HashValue
	RequestChainOutput *generic.OutputRef // output of 1i owned by the request originator
	Vars               map[string]string
}

func NewRequest(par NewRequestParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	tr := ret.Transfer()
	// create 1i transfer from RequestChainOutput to request account
	tr.AddInput(value.NewInputFromOutputRef(par.RequestChainOutput))
	chainOutIndex := tr.AddOutput(value.NewOutput(par.AssemblyAccount, 1))
	_, val := value.MustGetOutputAddrValue(par.RequestChainOutput)
	if val != 1 {
		return nil, fmt.Errorf("request chain output must have value exactly 1i")
	}
	reqBlk := sc.NewRequestBlock(par.AssemblyId, false).WithOutputIndices(chainOutIndex, 0, 0)

	vars := reqBlk.Vars()
	for k, v := range par.Vars {
		vars.SetString(generic.VarName(k), v)
	}
	ret.AddRequest(reqBlk)
	return ret, nil
}

func NewResultTransaction(reqRef *sc.RequestRef, config sc.Config) (sc.Transaction, error) {
	reqBlock := reqRef.RequestBlock()
	// check if request block points to valid chain output
	// which can be used as request->result chain
	requestChainOutputIdx, _, _ := reqBlock.OutputIndices()
	requestChainOutputRef := generic.NewOutputRef(reqRef.Tx().Transfer().Id(), requestChainOutputIdx)
	if !value.OutputCanBeChained(requestChainOutputRef, config.AssemblyAccount()) {
		return nil, fmt.Errorf("invalid request chain output")
	}
	tx := sc.NewTransaction()
	// add request chain link
	// transfer 1i from RequestChainAddress to itself
	tx.Transfer().AddInput(value.NewInputFromOutputRef(requestChainOutputRef))
	tx.Transfer().AddOutput(value.NewOutput(config.AssemblyAccount(), 1))
	return tx, nil
}

func NextStateUpdateTransaction(prevStateTx sc.Transaction, reqRef *sc.RequestRef) (sc.Transaction, error) {
	prevState, ok := prevStateTx.State()
	if !ok {
		return nil, fmt.Errorf("NextStateUpdateTransaction: state block not found")
	}
	tx, err := NewResultTransaction(reqRef, prevState.Config())
	if err != nil {
		return nil, err
	}
	tx.Transfer().AddInput(value.NewInput(prevStateTx.Transfer().Id(), prevStateTx.MustState().StateChainOutputIndex()))
	chainOutIdx := tx.Transfer().AddOutput(value.NewOutput(prevState.Config().AssemblyAccount(), 1))

	nextState := sc.NewStateBlock(prevState.AssemblyId(), prevState.Config().Id(), reqRef)
	nextState.
		WithStateIndex(prevState.StateIndex() + 1).
		WithVars(prevState.Vars()).
		WithStateChainOutputIndex(chainOutIdx)
	nextState.Config().With(prevState.Config())
	tx.SetState(nextState)
	return tx, nil
}

func ErrorTransaction(reqRef *sc.RequestRef, config sc.Config, err error) (sc.Transaction, error) {
	tx, err := NewResultTransaction(reqRef, config)
	if err != nil {
		return nil, err
	}
	errState := sc.NewStateBlock(reqRef.RequestBlock().AssemblyId(), config.Id(), reqRef).WithError(err)
	tx.SetState(errState)
	return tx, nil
}

// TODO sum is used for testing only

func SendOutputsToAddress(tx sc.Transaction, outputs []*generic.OutputRef, addr *HashValue, sum uint64) error {
	resum := uint64(0)
	for _, outp := range outputs {
		tx.Transfer().AddInput(value.NewInputFromOutputRef(outp))
		_, v := value.MustGetOutputAddrValue(outp)
		resum += v
	}
	if resum != sum {
		panic("resum != sum")
	}
	tx.Transfer().AddOutput(value.NewOutput(addr, sum))
	return nil
}
