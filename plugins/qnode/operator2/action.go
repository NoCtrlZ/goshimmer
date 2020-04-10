package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func (op *scOperator) takeAction() {
	op.doLeader()
	op.doSubordinate()
}

func (op *scOperator) doSubordinate() {
	for _, cr := range op.currentStateCompRequests {
		if cr.processed {
			continue
		}
		if cr.req.reqRef == nil {
			continue
		}
		cr.processed = true
		go op.processRequest(cr.req, cr.ts, cr.leaderPeerIndex)
	}
}

func (op *scOperator) doLeader() {
	if op.iAmCurrentLeader() {
		op.startProcessing()
	}
	if op.checkQuorum() {
		op.finalizeProcessing()
	}
}

func (op *scOperator) startProcessing() {
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
	msg := &startProcessingReqMsg{
		StateIndex: op.stateTx.MustState().StateIndex(),
		RequestId:  req.reqId,
	}
	var buf bytes.Buffer
	encodeProcessReqMsg(msg, &buf)
	numSucc, ts := op.comm.SendMsgToPeers(msgStartProcessingRequest, buf.Bytes())

	req.log.Debugf("%d 'msgStartProcessingRequest' messages sent to peers", numSucc)

	if numSucc < op.Quorum() {
		// doesn't make sense to continue because less than quorum sends succeeded
		req.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded", numSucc)
		return
	}
	op.leaderStatus = &leaderStatus{
		req:          req,
		ts:           ts,
		signedHashes: make([]signedHash, op.CommitteeSize()),
	}
	req.log.Debugf("msgStartProcessingRequest successfully sent to %d peers", numSucc)
	// run calculations async.
	go op.processRequest(req, ts, op.PeerIndex())
}

func (op *scOperator) checkQuorum() bool {
	log.Debug("checkQuorum")
	if op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized {
		log.Debug("checkQuorum: op.leaderStatus == nil || op.leaderStatus.resultTx == nil || op.leaderStatus.finalized")
		return false
	}
	mainHash := op.leaderStatus.signedHashes[op.PeerIndex()].MasterDataHash
	if mainHash == nil {
		log.Debug("checkQuorum: mainHash == nil")
		return false
	}
	quorumIndices := make([]int, 0, op.CommitteeSize())
	for i := range op.leaderStatus.signedHashes {
		if op.leaderStatus.signedHashes[i].MasterDataHash == nil {
			continue
		}
		if op.leaderStatus.signedHashes[i].MasterDataHash.Equal(mainHash) &&
			len(op.leaderStatus.signedHashes[i].SigBlocks) == len(op.leaderStatus.signedHashes[op.PeerIndex()].SigBlocks) {
			quorumIndices = append(quorumIndices, i)
		}
	}
	if len(quorumIndices) < int(op.Quorum()) {
		log.Debug("checkQuorum: len(quorumIndices) < int(op.Quorum())")
		return false
	}
	// quorum detected
	err := op.aggregateResult(quorumIndices, len(op.leaderStatus.signedHashes[op.PeerIndex()].SigBlocks))
	if err != nil {
		op.leaderStatus.req.log.Errorf("aggregateResult returned: %v", err)
		return false
	}
	err = sc.VerifySignatures(op.leaderStatus.resultTx, op.keyPool())
	if err != nil {
		op.leaderStatus.req.log.Errorf("VerifySignatures returned: %v", err)
		return false
	}
	return true
}

// takes action when stateChanged flag is true
func (op *scOperator) setNewState(tx sc.Transaction) {
	op.stateTx = tx
	// reset current leader seq index
	op.currLeaderSeqIndex = 0
	op.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), op.stateTx.Id().Bytes())
	for i, v := range op.leaderPeerIndexList {
		if v == op.PeerIndex() {
			op.myLeaderSeqIndex = uint16(i)
			break
		}
	}
	op.leaderStatus = nil

	op.currentStateCompRequests, op.nextStateCompRequests =
		op.nextStateCompRequests, op.currentStateCompRequests
	op.nextStateCompRequests = op.nextStateCompRequests[:0]

	op.requestNotificationsCurrentState, op.requestNotificationsNextState =
		op.requestNotificationsNextState, op.requestNotificationsCurrentState
	op.requestNotificationsNextState = op.requestNotificationsNextState[:0]

	// in the notification for the current state add all req ids from own queue of requests
	sortedReqs := op.sortedRequestsByAge()
	ids := make([]*sc.RequestId, len(sortedReqs))
	for i := range ids {
		ids[i] = sortedReqs[i].reqId
	}
	log.Debugf("setNewState: leader = %d iAmThLeader = %v", op.currentLeaderPeerIndex(), op.iAmCurrentLeader())

	op.accountRequestIdNotifications(op.PeerIndex(), op.stateTx.MustState().StateIndex(), ids...)
	// send notification about all requests to the current leader
	op.sendRequestNotificationsAllToLeader()
}

func (op *scOperator) selectRequestToProcess() *request {
	// vote
	votes := make(map[sc.RequestId]int)
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
			if req.reqRef != nil {
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

func (op *scOperator) iAmCurrentLeader() bool {
	return op.PeerIndex() == op.currentLeaderPeerIndex()
}

func (op *scOperator) currentLeaderPeerIndex() uint16 {
	return op.leaderPeerIndexList[op.currLeaderSeqIndex]
}
