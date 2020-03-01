package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

type runtimeContext struct {
	reqRef   *sc.RequestRef
	state    sc.Transaction
	resultTx sc.Transaction
	err      error
}

func (ctx *runtimeContext) RequestVars() generic.ValueMap {
	return ctx.reqRef.RequestBlock().Vars()
}

func (ctx *runtimeContext) StateVars() generic.ValueMap {
	return ctx.resultTx.MustState().Vars()
}

func (ctx *runtimeContext) ConfigVars() generic.ValueMap {
	return ctx.state.MustState().Config().Vars()
}

// BLS threshold signature. To use it as random value
func (ctx *runtimeContext) Signature() []byte {
	sig, _ := ctx.state.Signatures()[0].GetSignature()
	return sig
}

func (ctx *runtimeContext) SetError(err error) {
	ctx.err = err
}

func (ctx *runtimeContext) Error() error {
	return ctx.err
}

func (ctx *runtimeContext) RequestTransferId() *hashing.HashValue {
	return ctx.reqRef.Tx().Transfer().Id()
}

func (ctx *runtimeContext) MainRequestOutputs() [3]*generic.OutputRefWithValue {
	return ctx.reqRef.RequestBlock().MainOutputs(ctx.reqRef.Tx())
}

func (ctx *runtimeContext) SendFundsToAddress(outputs []*generic.OutputRefWithValue, addr *hashing.HashValue) {
	_ = clientapi.SendOutputsToAddress(ctx.resultTx, outputs, addr)
}

func (ctx *runtimeContext) AddRequestToSelf(reqNr uint16) {
	// add chain link to assembly account
	// select any unspent output which is not already used in the current result tx

}

// creates context with skeleton resulting transaction
// not signed

func newConfigUpdateRuntimeContext(reqRef *sc.RequestRef, curStateTx sc.Transaction) (*runtimeContext, error) {
	ownerAccount := curStateTx.MustState().Config().OwnerAccount()
	if !sc.AuthorizedForAddress(reqRef.Tx(), ownerAccount) {
		return nil, fmt.Errorf("config update request is not authorized")
	}

	resTx, err := clientapi.NextStateUpdateTransaction(curStateTx, reqRef)
	if err != nil {
		return nil, err
	}
	// just updates config variables
	resTx.MustState().WithVars(reqRef.RequestBlock().Vars())

	return &runtimeContext{
		reqRef:   reqRef,
		state:    curStateTx,
		resultTx: resTx,
	}, nil
}

func newStateUpdateRuntimeContext(reqRef *sc.RequestRef, curStateTx sc.Transaction) (*runtimeContext, error) {
	resTx, err := clientapi.NextStateUpdateTransaction(curStateTx, reqRef)
	if err != nil {
		return nil, err
	}
	return &runtimeContext{
		reqRef:   reqRef,
		state:    curStateTx,
		resultTx: resTx,
	}, nil
}
