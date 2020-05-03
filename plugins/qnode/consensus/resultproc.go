package consensus

import (
	"bytes"
	"time"
)

func (op *Operator) processRequest(req *request, ts time.Time, leaderPeerIndex uint16) {
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
	ctx.resultTx.MustState().WithTime(ctx.ts)

	req.log.Debugf("+++++  RES: %+v", ctx.resultTx.MustState().Vars())
	op.postEventToQueue(ctx)
}

func (op *Operator) sendResultToTheLeader(ctx *runtimeContext) {
	log.Debugw("sendResultToTheLeader",
		"req", ctx.reqRef.Id().Short(),
		"ts", ctx.ts,
	)
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

func (op *Operator) saveOwnResult(ctx *runtimeContext) {
	log.Debugw("saveOwnResult",
		"req", ctx.reqRef.Id().Short(),
		"ts", ctx.ts,
	)
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

// aggregates final signature, generates final result and posts to the tangle
func (op *Operator) finalizeProcessing() {
	req := op.leaderStatus.req
	vtx, err := op.leaderStatus.resultTx.ValueTx()
	if err != nil {
		req.log.Error(err)
		return
	}
	req.log.Debugw("POST result to the ValueTangle",
		"leader", op.PeerIndex(),
		"req", req.reqId.Short(),
		"resTx id", vtx.Id().Short())

	req.log.Info("FINALIZED REQUEST. Posting to the Value Tangle..")
	value.Post(vtx)
}

func (op *Operator) aggregateResult(quorumIndices []int, numSignatures int) error {
	resTx := op.leaderStatus.resultTx
	targetSigs, err := resTx.Signatures()
	if err != nil {
		return err
	}
	sigs := make([]generic.SignedBlock, 0, len(quorumIndices))
	for i := 0; i < numSignatures; i++ {
		for _, j := range quorumIndices {
			sigs = append(sigs, op.leaderStatus.signedHashes[j].SigBlocks[i])
		}
		err = generic.AggregateBLSBlocks(sigs, targetSigs[i], op.keyPool())
		if err != nil {
			return err
		}
		// verify recovered signature (testing)
		err = op.verifySignature(targetSigs[i])
		if err != nil {
			return err
		}
		sigs = sigs[:0]
	}
	return nil
}
