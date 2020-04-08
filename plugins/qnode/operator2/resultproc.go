package operator2

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func resultHash(stateIndex uint32, reqId, masterDataHash *HashValue) *HashValue {
	ret := HashData(
		tools.Uint32To4Bytes(stateIndex),
		reqId.Bytes(),
		masterDataHash.Bytes())
	//log.Debugf("+++++ resultHash (%d, %s, %s) -> %s",
	//	stateIndex, reqId.Short(), masterDataHash.Short(), ret.Short())
	return ret
}

func (op *scOperator) processRequest(leaderPeerIndex uint16) {
	req, ts := op.requestToProcess[0][leaderPeerIndex].req, op.requestToProcess[0][leaderPeerIndex].ts
	var ctx *runtimeContext
	var err error
	ctx, err = newStateUpdateRuntimeContext(leaderPeerIndex, req.reqRef, op.stateTx, ts)
	if err != nil {
		req.log.Warnw("can't create runtime context",
			"aid", req.reqRef.RequestBlock().SContractId().Short(),
			"isConfigUpdate", req.reqRef.RequestBlock().IsConfigUpdateReq(),
			"err", err,
		)
		return
	}
	op.processor.Run(ctx)
	displayResult(req, ctx)
	op.postEventToQueue(ctx)
}

func displayResult(req *request, ctx *runtimeContext) {
	req.log.Debugf("+++++  RES: %+v", ctx.resultTx.MustState().Vars())
}

func (op *scOperator) sendResultToTheLeader(leaderPeerIndex uint16) {
	resultTx := op.requestToProcess[0][leaderPeerIndex].ownResult
	err := sc.SignTransaction(resultTx, op.keyPool())
	if err != nil {
		op.requestToProcess[0][leaderPeerIndex].req.log.Errorf("SignTransaction returned: %v", err)
		return
	}
	sigs, err := resultTx.Signatures()
	if err != nil {
		op.requestToProcess[0][leaderPeerIndex].req.log.Error(err)
		return
	}
	msg := &signedHashMsg{
		StateIndex:    op.stateTx.MustState().StateIndex(),
		RequestId:     op.requestToProcess[0][leaderPeerIndex].reqId,
		OrigTimestamp: op.requestToProcess[0][leaderPeerIndex].ts,
		DataHash:      resultTx.MasterDataHash(),
		SigBlocks:     sigs,
	}
	var buf bytes.Buffer
	encodeSignedHashMsg(msg, &buf)
	if err := op.comm.SendMsg(leaderPeerIndex, msgSignedHash, buf.Bytes()); err != nil {
		op.requestToProcess[0][leaderPeerIndex].req.log.Error(err)
	}
}
