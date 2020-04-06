package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
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

func (op *scOperator) asyncCalculateResult(req *request, ts time.Time) {
	if op.currentResult != nil {
		return
	}
	if req.reqRef == nil {
		return
	}
	taskId := HashData(req.reqId.Bytes(), op.stateTx.Id().Bytes())
	if _, ok := req.startedCalculation[*taskId]; !ok {
		req.startedCalculation[*taskId] = time.Now()
		req.log.Debugf("start calculation in state idx %d", op.stateTx.MustState().StateIndex())
		go op.processRequest(req, ts)
	}
}

func (op *scOperator) processRequest(req *request, ts time.Time) {
	var ctx *runtimeContext
	var err error
	if req.reqRef.RequestBlock().IsConfigUpdateReq() {
		ctx, err = newConfigUpdateRuntimeContext(req.reqRef, op.stateTx, ts)
	} else {
		ctx, err = newStateUpdateRuntimeContext(req.reqRef, op.stateTx, ts)
	}
	if err != nil {
		req.log.Warnw("can't create runtime context",
			"aid", req.reqRef.RequestBlock().SContractId().Short(),
			"isConfigUpdate", req.reqRef.RequestBlock().IsConfigUpdateReq(),
			"err", err,
		)
		return
	}
	if !req.reqRef.RequestBlock().IsConfigUpdateReq() {
		// non config updates are passed to processor
		op.processor.Run(ctx)
		displayResult(req, ctx)
	}
	op.postEventToQueue(ctx)
}

func displayResult(req *request, ctx *runtimeContext) {
	req.log.Debugf("+++++  RES: %+v", ctx.resultTx.MustState().Vars())
}

func (op *scOperator) pushResultMsgFromResult(resRec *resultCalculated) (*pushResultMsg, error) {
	sigBlocks, err := resRec.res.resultTx.Signatures()
	if err != nil {
		return nil, err
	}
	state, _ := resRec.res.state.State()
	return &pushResultMsg{
		SenderIndex:    op.PeerIndex(),
		RequestId:      resRec.res.reqRef.Id(),
		StateIndex:     state.StateIndex(),
		MasterDataHash: resRec.masterDataHash,
		SigBlocks:      sigBlocks,
	}, nil
}

func (op *scOperator) sendPushResultToPeer(res *resultCalculated, peerIndex uint16) {
	locLog := log
	if req, ok := op.requestFromId(res.res.reqRef.Id()); ok {
		locLog = req.log
	}

	pushMsg, err := op.pushResultMsgFromResult(res)
	if err != nil {
		locLog.Errorf("sendPushResultToPeer: %v", err)
		return
	}

	resultHash := resultHash(pushMsg.StateIndex, pushMsg.RequestId, pushMsg.MasterDataHash)
	locLog.Debugf("sendPushResultToPeer %d for state idx %d, res hash %s",
		peerIndex, res.res.state.MustState().StateIndex(), resultHash.Short())

	var encodedMsg bytes.Buffer
	encodePushResultMsg(pushMsg, &encodedMsg)
	err = op.comm.SendMsg(peerIndex, msgTypePush, encodedMsg.Bytes())
	if err != nil {
		locLog.Errorf("SendUDPData returned error: `%v`", err)
	}
}
