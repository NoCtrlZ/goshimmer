package operator2

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

// aggregates final signature, generates final result and posts to the tangle

func (op *scOperator) finalizeProcessing() {
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

func (op *scOperator) aggregateResult(quorumIndices []int, numSignatures int) error {
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
