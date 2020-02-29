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

// return deposit output index and value

func (ctx *runtimeContext) GetDepositOutput() (uint16, uint64) {
	reqBlock := ctx.reqRef.RequestBlock()
	_, _, depoIdx := reqBlock.OutputIndices()
	depoOut := ctx.reqRef.Tx().Transfer().Outputs()[depoIdx]
	return depoIdx, depoOut.Value()
}

func (ctx *runtimeContext) SendFundsToAddress(outputs []*generic.OutputRef, addr *hashing.HashValue, value uint64) {
	_ = clientapi.SendOutputsToAddress(ctx.resultTx, outputs, addr, value)
}

func (ctx *runtimeContext) AddRequestToSelf(reqNr uint16) {
	//ctx.resultTx.Transfer().AddInput(value.NewInputFromOutputRef(outRef))
	//return tr.AddOutput(NewOutput(addr, 1))
	// TODO

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
