package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"time"
)

func (op *AssemblyOperator) finalizeTheRequest(res *resultCalculated) error {
	// aggregates final signature, generates final result and posts to the tangle
	err := op.aggregateResult(res)
	if err != nil {
		return err
	}
	log.Infow("finalizeTheRequest POST", "peer", op.peerIndex(), "res tx id", res.res.resultTx.Id().Short())
	res.finalized = true
	res.finalizedWhen = time.Now()
	vtx, err := res.res.resultTx.ValueTx()
	if err != nil {
		return err
	}
	return op.comm.PostToValueTangle(vtx)
}

func (op *AssemblyOperator) aggregateResult(res *resultCalculated) error {
	reqId := res.res.reqRef.Id()
	reqRec, _ := op.requestFromId(reqId)
	rhlst, ok := reqRec.receivedResultHashes[*res.rsHash]
	if !ok {
		log.Panic("aggregateResult: inconsistency: no shares found")
	}
	numNotNil := uint16(0)
	for _, rh := range rhlst {
		if rh != nil {
			numNotNil++
		}
	}
	if numNotNil+1 < op.requiredQuorum() {
		// must be checked before
		log.Panic("aggregateResult: inconsistency: not enough shares to finalize result")
	}

	ownSignedBlocks := reqRec.ownResultCalculated.res.resultTx.Signatures()

	for i, sigBlock := range ownSignedBlocks {
		receivedSigBlocks := make([]generic.SignedBlock, 0, op.assemblySize())
		for _, rh := range rhlst {
			if len(ownSignedBlocks) != len(rh.SigBlocks) {
				return fmt.Errorf("unexpected different lengths of signature lists")
			}
			sb := rh.SigBlocks[i]
			receivedSigBlocks = append(receivedSigBlocks, sb)
		}
		err := op.aggregateBlocks(receivedSigBlocks, sigBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
