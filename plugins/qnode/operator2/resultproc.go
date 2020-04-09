package operator2

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
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

func (op *scOperator) processRequest(req *request, ts time.Time, leaderPeerIndex uint16) {
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

func (op *scOperator) sendResultToTheLeader(ctx *runtimeContext) {
	req, _ := op.requestFromId(ctx.reqRef.Id())
	err := sc.SignTransaction(ctx.resultTx, op.keyPool())
	if err != nil {
		req.log.Errorf("SignTransaction returned: %v", err)
		return
	}
	sigs, err := ctx.resultTx.Signatures()
	if err != nil {
		req.log.Errorf("Signatures returned: %v", err)
		return
	}
	msg := &signedHashMsg{
		StateIndex:    op.stateTx.MustState().StateIndex(),
		RequestId:     ctx.reqRef.Id(),
		OrigTimestamp: ctx.Time(),
		DataHash:      ctx.resultTx.MasterDataHash(),
		SigBlocks:     sigs,
	}
	var buf bytes.Buffer
	encodeSignedHashMsg(msg, &buf)
	if err := op.comm.SendMsg(ctx.leaderIndex, msgSignedHash, buf.Bytes()); err != nil {
		req.log.Error(err)
	}
}

func (op *scOperator) saveOwnResult(ctx *runtimeContext) {
	req, _ := op.requestFromId(ctx.reqRef.Id())
	err := sc.SignTransaction(ctx.resultTx, op.keyPool())
	if err != nil {
		req.log.Errorf("SignTransaction returned: %v", err)
		return
	}
	sigs, err := ctx.resultTx.Signatures()
	if err != nil {
		req.log.Errorf("Signatures returned: %v", err)
		return
	}
	op.leaderStatus.resultTx = ctx.resultTx
	op.leaderStatus.signedHashes[op.PeerIndex()].MasterDataHash = ctx.resultTx.MasterDataHash()
	op.leaderStatus.signedHashes[op.PeerIndex()].SigBlocks = sigs

}
