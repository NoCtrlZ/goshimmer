package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

type runtimeContext struct {
	reqRef   *sc.RequestRef
	state    sc.Transaction
	resultTx sc.Transaction
}

func (ctx *runtimeContext) InputVars() generic.ValueMap {
	return ctx.reqRef.RequestBlock().Vars()
}

func (ctx *runtimeContext) OutputVars() generic.ValueMap {
	return ctx.resultTx.MustState().Vars()
}

// creates context with skeleton resulting transaction
// not signed

func newConfigUpdateRuntimeContext(reqRef *sc.RequestRef, curStateTx sc.Transaction) (*runtimeContext, error) {
	ownerAccount := curStateTx.MustState().Config().OwnerAccount()
	if !sc.AuthorizedForAddress(reqRef.Tx(), ownerAccount) {
		return nil, fmt.Errorf("config update request is not authorized")
	}

	resTx, err := sc.NextStateUpdateTransaction(curStateTx, reqRef)
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
	resTx, err := sc.NextStateUpdateTransaction(curStateTx, reqRef)
	if err != nil {
		return nil, err
	}
	return &runtimeContext{
		reqRef:   reqRef,
		state:    curStateTx,
		resultTx: resTx,
	}, nil
}
