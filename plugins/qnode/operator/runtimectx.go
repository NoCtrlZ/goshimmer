package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/hive.go/logger"
)

type runtimeContext struct {
	reqRef   *sc.RequestRef
	state    sc.Transaction
	resultTx sc.Transaction
	log      *logger.Logger
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

func (ctx *runtimeContext) AssemblyAccount() *hashing.HashValue {
	return ctx.state.MustState().Config().SContractAccount()
}

// BLS threshold signature. To use it as random value
func (ctx *runtimeContext) Signature() []byte {
	sigs, err := ctx.state.Signatures()
	if err != nil {
		return nil
	}
	sig, _ := sigs[0].GetSignature()
	return sig
}

func (ctx *runtimeContext) SetError(err error) {
	if ctx.log != nil {
		ctx.log.Errorf("SetError: %v", err)
	}
	ctx.err = err
}

func (ctx *runtimeContext) Error() error {
	return ctx.err
}

func (ctx *runtimeContext) RequestTransferId() *hashing.HashValue {
	return ctx.reqRef.Tx().Transfer().Id()
}

func (ctx *runtimeContext) MainRequestOutputs() sc.MainRequestOutputs {
	return ctx.reqRef.RequestBlock().MainOutputs(ctx.reqRef.Tx())
}

func (ctx *runtimeContext) MainInputAddress() (*hashing.HashValue, error) {
	tr := ctx.reqRef.Tx().Transfer()
	if len(tr.Inputs()) == 0 {
		return nil, fmt.Errorf("no inputs. Can't find main input address") // panic?
	}
	inp0 := tr.Inputs()[0].OutputRef()
	outp, err := value.GetOutputAddrValue(inp0)
	if err != nil {
		return nil, fmt.Errorf("can't find main input address") // panic?

	}
	return outp.Addr, nil
}

func (ctx *runtimeContext) SendFundsToAddress(outputs []*generic.OutputRef, addr *hashing.HashValue) {
	_ = clientapi.SendAllOutputsToAddress(ctx.resultTx, outputs, addr)
}

func (ctx *runtimeContext) SendOutputsToOutputs(inOutputs []*generic.OutputRef, outOutputs []value.Output, reminderAddr *hashing.HashValue) error {
	return clientapi.SendOutputsToOutputs(ctx.resultTx, inOutputs, outOutputs, reminderAddr)
}

func (ctx *runtimeContext) AddRequestToSelf(reqType uint16) error {
	vars := generic.NewFlatValueMap()
	vars.SetInt("req_type", int(reqType))
	_, err := clientapi.AddNewRequestBlock(ctx.resultTx, clientapi.NewRequestParams{
		AssemblyId:       ctx.state.MustState().SContractId(),
		AssemblyAccount:  ctx.AssemblyAccount(),
		RequesterAccount: ctx.AssemblyAccount(),
		Vars:             vars,
	})
	return err
}

func (ctx *runtimeContext) Log() *logger.Logger {
	return ctx.log
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
		log:      logger.NewLogger("VM"),
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
		log:      logger.NewLogger("VM"),
	}, nil
}
