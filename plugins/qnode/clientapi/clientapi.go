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
	OriginOutput     *generic.OutputRef // output of 1i owned by the owner
	OwnersPrivateKey interface{}        // owner's private key data used to sign the input
}

const (
	MAP_KEY_STATE_ACCOUNT   = "state_addr"
	MAP_KEY_REQUEST_ACCOUNT = "request_addr"
	MAP_KEY_OWNER_ACCOUNT   = "owner_addr"
)

func NewOriginTransaction(par NewOriginParams) (sc.Transaction, error) {
	ret := sc.NewTransaction()
	state := sc.NewStateBlock(par.AssemblyId, par.ConfigId, NilHash, 0)
	configVars := state.ConfigVars()
	configVars.SetString(MAP_KEY_STATE_ACCOUNT, par.StateAccount.String())
	configVars.SetString(MAP_KEY_REQUEST_ACCOUNT, par.RequestAccount.String())
	configVars.SetString(MAP_KEY_OWNER_ACCOUNT, par.OwnerAccount.String())
	ret.SetState(state)
	tr := ret.Transfer()

	// adding owner chain: transfer of 1i from owner's account to the stateAccount
	// the latter will be used to build chain
	addr, val := value.GetAddrValue(par.OriginOutput)
	if !addr.Equal(par.OwnerAccount) || val != 1 {
		return nil, fmt.Errorf("OriginOutput must be exactly 1i from owner's account")
	}
	tr.AddInput(value.NewInputFromOutputRef(par.OriginOutput))
	tr.AddOutput(value.NewOutput(par.StateAccount, 1))
	// signing inputs with the owner's private key
	sigs := tr.InputSignatures()
	sig, ok := sigs[*par.OwnerAccount]
	if !ok {
		panic("too bad")
	}
	sig.SetSignature(NilHash.Bytes(), generic.SIG_TYPE_FAKE) // fake. must be signed by the owner
	return ret, nil
}
