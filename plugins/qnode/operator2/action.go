package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func (op *scOperator) takeAction() {
	op.doLeader()
}

func (op *scOperator) doLeader() {
	// when operator is rotated to the leader position,
	// the 'leader' flag is up and doesn't change since
	// the meaning is: the operator has been doing its job as the leader of the current state
	op.requestToProcess[0][op.PeerIndex()].leader = op.requestToProcess[0][op.PeerIndex()].leader || op.iAmCurrentLeader()
	if !op.iAmCurrentLeader() {
		op.startProcessing()
	}
	if op.requestToProcess[0][op.PeerIndex()].leader && !op.requestToProcess[0][op.PeerIndex()].finalized {
		if op.checkQuorum() {
			op.finalizeProcessing()
		}
	}
}

func (op *scOperator) startProcessing() {
	if op.requestToProcess[0][op.PeerIndex()].req != nil {
		// request already selected and calculations initialized
		return
	}
	req := op.selectRequestToProcess()
	if req == nil {
		// can't select request to process
		return
	}
	msg := &startProcessingReqMsg{
		StateIndex: op.stateTx.MustState().StateIndex(),
		RequestId:  req.reqId,
	}
	var buf bytes.Buffer
	encodeProcessReqMsg(msg, &buf)
	numSucc, ts := op.comm.SendMsgToPeers(msgStartProcessingRequest, buf.Bytes())

	if numSucc < op.Quorum() {
		// doesn't make sense to continue because less than quorum sends succeeded
		req.log.Errorf("only %d 'msgStartProcessingRequest' sends succeeded", numSucc)
		return
	}
	op.requestToProcess[0][op.PeerIndex()].req = req
	op.requestToProcess[0][op.PeerIndex()].ts = ts

	req.log.Debugf("msgStartProcessingRequest successfully sent to %d peers", numSucc)
	// run calculations async.
	go op.processRequest(op.PeerIndex())
}

func (op *scOperator) checkQuorum() bool {
	if !op.requestToProcess[0][op.PeerIndex()].leader {
		return false
	}
	mainHash := op.requestToProcess[0][op.PeerIndex()].MasterDataHash
	if mainHash == nil || op.requestToProcess[0][op.PeerIndex()].ownResult == nil {
		return false
	}
	quorumIndices := make([]int, 0, op.CommitteeSize())
	for i := range op.requestToProcess[0] {
		if op.requestToProcess[0][i].MasterDataHash == nil {
			continue
		}
		if op.requestToProcess[0][i].MasterDataHash.Equal(mainHash) &&
			len(op.requestToProcess[0][i].SigBlocks) == len(op.requestToProcess[0][op.PeerIndex()].SigBlocks) {
			quorumIndices = append(quorumIndices, i)
		}
	}
	if len(quorumIndices) < int(op.Quorum()) {
		return false
	}
	// quorum detected
	err := op.aggregateResult(quorumIndices, len(op.requestToProcess[0][op.PeerIndex()].SigBlocks))
	if err != nil {
		op.requestToProcess[0][op.PeerIndex()].req.log.Errorf("aggregateResult returned: %v", err)
		return false
	}
	err = sc.VerifySignatures(op.requestToProcess[0][op.PeerIndex()].ownResult, op.keyPool())
	if err != nil {
		op.requestToProcess[0][op.PeerIndex()].req.log.Errorf("VerifySignatures returned: %v", err)
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

	// swap arrays of incoming initReq's
	// clean the array of the next state
	op.requestToProcess[0], op.requestToProcess[1] = op.requestToProcess[1], op.requestToProcess[0]
	for i := range op.requestToProcess[1] {
		op.requestToProcess[1][i] = processingStatus{}
	}
	// swap curr and next state request notifications for each peer
	// clean the notifications for the next state index
	for i := range op.requestNotificationsReceived {
		op.requestNotificationsReceived[i][0], op.requestNotificationsReceived[i][1] =
			op.requestNotificationsReceived[i][1], op.requestNotificationsReceived[i][0]
		op.requestNotificationsReceived[i][1] = op.requestNotificationsReceived[i][1][:0]
	}
	// in the notification for the current state add all req ids from own queue of requests
	sortedReqs := op.sortedRequestsByAge()
	ids := make([]*sc.RequestId, len(sortedReqs))
	for i := range ids {
		ids[i] = sortedReqs[i].reqId
	}
	op.accountRequestIdNotifications(op.PeerIndex(), false, ids...)

	// send notification about all requests to the current leader
	op.sendRequestNotificationsAllToLeader()
}
