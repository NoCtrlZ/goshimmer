package consensus

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

func (op *Operator) takeAction() {
	op.doLeader()
	op.doSubordinate()
}

func (op *Operator) doSubordinate() {
	for _, cr := range op.currentStateCompRequests {
		if cr.processed {
			continue
		}
		if cr.req.reqMsg == nil {
			continue
		}
		cr.processed = true
		//go op.processRequest(cr.req, cr.ts, cr.leaderPeerIndex)
	}
}

func (op *Operator) doLeader() {
	if op.iAmCurrentLeader() {
		op.startProcessing()
	}
	if op.checkQuorum() {
		//op.finalizeProcessing()
	}
}

func (op *Operator) startProcessing() {
	if op.leaderStatus != nil {
		// request already selected and calculations initialized
		return
	}
	req := op.selectRequestToProcess()
	if req == nil {
		// can't select request to process
		log.Debugf("can't select request to process")
		return
	}
	req.log.Debugw("request selected to process", "stateIdx", op.stateTx.MustState().StateIndex())
	msgData := hashing.MustBytes(&committee.StartProcessingReqMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: op.stateTx.MustState().StateIndex(),
		},
		RequestId: req.reqId,
	})

	numSucc, ts := op.committee.SendMsgToPeers(committee.MsgStartProcessingRequest, msgData)

	req.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.Quorum()-1 {
		// doesn't make sense to continue because less than quorum sends succeeded
		req.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded", numSucc)
		return
	}
	op.leaderStatus = &leaderStatus{
		req:          req,
		ts:           ts,
		signedHashes: make([]signedHash, op.committee.Size()),
	}
	req.log.Debugf("msgStartProcessingRequest successfully sent to %d peers", numSucc)
	// run calculations async.
	//go op.processRequest(req, ts, op.PeerIndex())
}

func (op *Operator) checkQuorum() bool {
	log.Debug("checkQuorum")
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		//log.Debug("checkQuorum: op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized")
		return false
	}
	mainHash := op.leaderStatus.signedHashes[op.committee.OwnPeerIndex()].MasterDataHash
	if mainHash == nil {
		//log.Debug("checkQuorum: mainHash == nil")
		return false
	}
	quorumIndices := make([]int, 0, op.committee.Size())
	for i := range op.leaderStatus.signedHashes {
		if op.leaderStatus.signedHashes[i].MasterDataHash == nil {
			continue
		}
		//if op.leaderStatus.signedHashes[i].MasterDataHash.Equal(mainHash) &&
		//	len(op.leaderStatus.signedHashes[i].SigBlocks) == len(op.leaderStatus.signedHashes[op.PeerIndex()].SigBlocks) {
		//	quorumIndices = append(quorumIndices, i)
		//}
	}
	if len(quorumIndices) < int(op.Quorum()) {
		//log.Debug("checkQuorum: len(quorumIndices) < int(op.Quorum())")
		return false
	}
	// quorum detected
	//err := op.aggregateResult(quorumIndices, len(op.leaderStatus.signedHashes[op.PeerIndex()].SigBlocks))
	//if err != nil {
	//	op.leaderStatus.req.log.Errorf("aggregateResult returned: %v", err)
	//	return false
	//}
	//err = sc.VerifySignatures(op.leaderStatus.resultTx, op.keyPool())
	//if err != nil {
	//	op.leaderStatus.req.log.Errorf("VerifySignatures returned: %v", err)
	//	return false
	//}
	return true
}

// sets new state transaction and initializes respective variables
func (op *Operator) setNewState(stateTx *sctransaction.Transaction, variableState state.VariableState) {
	op.stateTx = stateTx
	op.variableState = variableState
	op.resetLeader()

	// computation requests and notifications about requests for the next state index
	// are brought to the current state next state list is cleared
	op.currentStateCompRequests, op.nextStateCompRequests =
		op.nextStateCompRequests, op.currentStateCompRequests
	op.nextStateCompRequests = op.nextStateCompRequests[:0]

	op.requestNotificationsCurrentState, op.requestNotificationsNextState =
		op.requestNotificationsNextState, op.requestNotificationsCurrentState
	op.requestNotificationsNextState = op.requestNotificationsNextState[:0]
}

func (op *Operator) selectRequestToProcess() *request {
	// vote
	votes := make(map[sctransaction.RequestId]int)
	for _, rn := range op.requestNotificationsCurrentState {
		if _, ok := votes[*rn.reqId]; !ok {
			votes[*rn.reqId] = 0
		}
		votes[*rn.reqId] = votes[*rn.reqId] + 1
	}
	if len(votes) == 0 {
		return nil
	}
	maxvotes := 0
	for _, v := range votes {
		if v > maxvotes {
			maxvotes = v
		}
	}
	if maxvotes < int(op.Quorum()) {
		return nil
	}
	candidates := make([]*request, 0, len(votes))
	for rid, v := range votes {
		if v == int(op.Quorum()) {
			req := op.requests[rid]
			if req.reqMsg != nil {
				candidates = append(candidates, req)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sortRequestsByAge(candidates)
	return candidates[0]
}
