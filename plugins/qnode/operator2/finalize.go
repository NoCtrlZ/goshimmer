package operator2

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

// aggregates final signature, generates final result and posts to the tangle

func (op *scOperator) finalizeProcessing() {
	req := op.requestToProcess[0][op.PeerIndex()].req
	vtx, err := op.requestToProcess[0][op.PeerIndex()].ownResult.ValueTx()
	if err != nil {
		req.log.Error(err)
		return
	}
	req.log.Debugw("POST result to the ValueTangle",
		"leader", op.PeerIndex(),
		"req", req.reqId.Short(),
		"resTx id", op.requestToProcess[0][op.PeerIndex()].ownResult.Id().Short())

	req.log.Info("FINALIZED REQUEST. Posting to the Value Tangle..")
	value.Post(vtx)
}

func (op *scOperator) aggregateResult(quorumIndices []int, numSignatures int) error {
	resTx := op.requestToProcess[0][op.PeerIndex()].ownResult
	targetSigs, err := resTx.Signatures()
	if err != nil {
		return err
	}

	sigs := make([]generic.SignedBlock, len(quorumIndices))
	for i := 0; i < numSignatures; i++ {
		for _, j := range quorumIndices {
			sigs[j] = op.requestToProcess[0][j].SigBlocks[i]
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
	}
	return nil
}
