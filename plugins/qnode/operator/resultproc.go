package operator

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"time"
)

func isCalculationInProgress(req *request, resHash *HashValue) bool {
	_, ok := req.startedCalculation[*resHash]
	return ok
}

func markCalculationInProgress(req *request, resHash *HashValue) {
	req.startedCalculation[*resHash] = time.Now()
}

func (op *AssemblyOperator) asyncCalculateResult(req *request) {
	if req.ownResultCalculated != nil {
		return
	}
	if req.reqRef == nil {
		return
	}
	rsHash := HashData(req.reqId.Bytes(), op.stateTx.Id().Bytes())
	if isCalculationInProgress(req, rsHash) {
		// already started
		return
	}
	markCalculationInProgress(req, rsHash)
	go op.processRequest(req)
}

func (op *AssemblyOperator) processRequest(req *request) {
	var ctx *runtimeContext
	var err error
	if req.reqRef.RequestBlock().IsConfigUpdateReq() {
		ctx, err = newConfigUpdateRuntimeContext(req.reqRef, op.stateTx)
	} else {
		ctx, err = newStateUpdateRuntimeContext(req.reqRef, op.stateTx)
	}
	if err != nil {
		log.Warnw("can't create runtime context",
			"aid", req.reqRef.RequestBlock().AssemblyId().Short(),
			"req tx", req.reqRef.Tx().Id(),
			"req id", req.reqRef.Id(),
			"isConfigUpdate", req.reqRef.RequestBlock().IsConfigUpdateReq(),
			"err", err,
		)
		return
	}
	if !req.reqRef.RequestBlock().IsConfigUpdateReq() {
		// non config updates are passed to processor
		op.processor.Run(ctx)
	}
	op.DispatchEvent(ctx)
}

func (op *AssemblyOperator) pushResultMsgFromResult(resRec *resultCalculated) *pushResultMsg {
	sigBlocks := resRec.res.resultTx.Signatures()
	state, _ := resRec.res.state.State()
	return &pushResultMsg{
		SenderIndex:    op.peerIndex(),
		RequestId:      resRec.res.reqRef.Id(),
		StateIndex:     state.StateIndex(),
		MasterDataHash: resRec.masterDataHash,
		SigBlocks:      sigBlocks,
	}
}

func (op *AssemblyOperator) sendPushResultToPeer(res *resultCalculated, peerIndex uint16) {
	data, _ := op.encodeMsg(op.pushResultMsgFromResult(res))

	if peerIndex == op.peerIndex() {
		log.Error("error: attempt to send result hash to itself. Result hash wasn't sent")
		return
	}
	addr := op.peers[peerIndex]
	err := op.comm.SendUDPData(data, op.assemblyId, op.peerIndex(), MSG_RESULT_HASH, addr)
	if err != nil {
		log.Errorf("SendUDPData returned error: `%v`", err)
	}
}
