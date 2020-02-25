package operator

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func resultHash(stateIndex uint32, reqId, masterDataHash *HashValue) *HashValue {
	return HashData(
		tools.Uint32To4Bytes(stateIndex),
		reqId.Bytes(),
		masterDataHash.Bytes())
}

func (op *AssemblyOperator) asyncCalculateResult(req *request) {
	if req.ownResultCalculated != nil {
		return
	}
	if req.reqRef == nil {
		return
	}
	taskId := HashData(req.reqId.Bytes(), op.stateTx.Id().Bytes())
	if _, ok := req.startedCalculation[*taskId]; !ok {
		req.startedCalculation[*taskId] = time.Now()
		log.Debugw("start calculation",
			"req", req.reqId.Short(),
			"state idx", op.stateTx.MustState().StateIndex(),
		)
		go op.processRequest(req)
	}
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
	log.Debugw("sendPushResultToPeer",
		"peer", peerIndex,
		"req", res.res.reqRef.Id().Short(),
		"state idx", res.res.state.MustState().StateIndex(),
	)
	data, _ := op.encodeMsg(op.pushResultMsgFromResult(res))

	if peerIndex == op.peerIndex() {
		log.Error("error: attempt to send result hash to itself. Result hash wasn't sent")
		return
	}
	addr := op.peers[peerIndex]
	err := op.comm.SendUDPData(data, op.assemblyId, op.peerIndex(), MSG_PUSH_MSG, addr)
	if err != nil {
		log.Errorf("SendUDPData returned error: `%v`", err)
	}
}
