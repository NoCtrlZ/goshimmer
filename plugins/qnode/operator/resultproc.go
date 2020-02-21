package operator

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/modelimpl"
	"time"
)

func isCalculationInProgress(req *request, resHash *HashValue) bool {
	_, ok := req.startedCalculation[*resHash]
	return ok
}

func markCalculationInProgress(req *request, resHash *HashValue) {
	req.startedCalculation[*resHash] = time.Now()
}

func (op *AssemblyOperator) processRequest(req *request) {
	reqBlock := req.msgTx.Requests()[req.msgIndex]
	if reqBlock.IsConfigUpdateReq() {
		op.processConfigUpdateRequest(req)
	} else {
		op.asyncCalculateResult(req)
	}
}

func (op *AssemblyOperator) processConfigUpdateRequest(req *request) {
	// check owner authorisation
	ownerAddr := op.stateTx.MustState().OwnerChainAddress()
	if !modelimpl.AuthorizedForAddress(req.msgTx.Transfer(), ownerAddr) {
		return
	}
	// config update request
	nextState, _ := newResultCalculated(req.msgTx, req.msgIndex, op.stateTx)
	reqBlock := req.msgTx.Requests()[req.msgIndex]
	nextState.resultTx.MustState().Builder().SetConfigVars(reqBlock.Vars())

	// update state synchronously
	op.EventStateUpdate(nextState.resultTx)
}

func (op *AssemblyOperator) asyncCalculateResult(req *request) {
	if req.ownResultCalculated != nil {
		return
	}
	if req.msgTx == nil {
		return
	}
	rsHash := HashData(req.reqId.Bytes(), op.stateTx.Id().Bytes())
	if isCalculationInProgress(req, rsHash) {
		// already started
		return
	}

	markCalculationInProgress(req, rsHash)
	go func() {
		ctx, _ := newResultCalculated(req.msgTx, req.msgIndex, op.stateTx)
		op.processor.Run(ctx)
		op.DispatchEvent(ctx)
	}()
}

func (op *AssemblyOperator) pushResultMsgFromResult(resRec *resultCalculatedIntern) *pushResultMsg {
	sigBlocks := resRec.res.resultTx.Signatures()
	state, _ := resRec.res.state.State()
	return &pushResultMsg{
		SenderIndex:    op.peerIndex(),
		RequestId:      resRec.res.requestTx.Id(),
		StateIndex:     state.StateIndex(),
		MasterDataHash: resRec.masterDataHash,
		SigBlocks:      sigBlocks,
	}
}

func (op *AssemblyOperator) sendPushResultToPeer(res *resultCalculatedIntern, peerIndex uint16) {
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
