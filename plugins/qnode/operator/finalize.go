package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
)

func maxVotesFromPeers(req *request) (uint16, *HashValue) {
	var retRsHash HashValue
	var retNumVotes uint16

	for rsHash, rhlst := range req.pushMessages {
		numNotNil := uint16(0)
		for _, rh := range rhlst {
			if rh != nil {
				numNotNil++
			}
		}
		if numNotNil > retNumVotes {
			retNumVotes = numNotNil
			copy(retRsHash.Bytes(), rsHash.Bytes())
		}
	}
	return retNumVotes, &retRsHash
}

func (op *AssemblyOperator) finalizeTheRequest(res *resultCalculated) error {
	// aggregates final signature, generates final result and posts to the tangle
	err := op.aggregateResult(res)
	if err != nil {
		return err
	}
	err = sc.VerifySignedBlocks(res.res.resultTx.Signatures(), op)
	if err != nil {
		return err
	}

	res.finalized = true
	res.finalizedWhen = time.Now()
	vtx, err := res.res.resultTx.ValueTx()
	if err != nil {
		return err
	}
	log.Infow("POST result to the ValueTangle",
		"leader", op.peerIndex(),
		"req", res.res.reqRef.Id().Short(),
		"resTx id", res.res.resultTx.Id().Short())

	return op.comm.PostToValueTangle(vtx)
}

func (op *AssemblyOperator) aggregateResult(res *resultCalculated) error {
	reqId := res.res.reqRef.Id()
	reqRec, ok := op.requestFromId(reqId)
	if !ok {
		log.Panic("aggregateResult: no request found")
	}
	rhlst, ok := reqRec.pushMessages[*res.resultHash]
	if !ok {
		log.Panic("aggregateResult: inconsistency: no shares found")
	}
	numNotNil := uint16(0)
	for _, rh := range rhlst {
		if rh != nil {
			numNotNil++
		}
	}
	if numNotNil+1 < op.assemblyQuorum() {
		// must be checked before
		log.Panic("aggregateResult: inconsistency: not enough shares to finalize result")
	}
	ownSignedBlocks := reqRec.ownResultCalculated.res.resultTx.Signatures()

	for i, sigBlock := range ownSignedBlocks {
		receivedSigBlocks := make([]generic.SignedBlock, 0, op.assemblySize())
		for _, rh := range rhlst {
			if rh == nil {
				continue
			}
			if len(ownSignedBlocks) != len(rh.SigBlocks) {
				return fmt.Errorf("unexpected different lengths of signature lists")
			}
			receivedSigBlocks = append(receivedSigBlocks, rh.SigBlocks[i])
		}
		receivedSigBlocks = append(receivedSigBlocks, sigBlock)
		err := generic.AggregateBLSBlocks(receivedSigBlocks, sigBlock, op)
		if err != nil {
			return err
		}
		// verify recovered signature (testing)
		err = op.VerifySignature(sigBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
