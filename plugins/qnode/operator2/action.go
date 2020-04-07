package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func (op *scOperator) takeAction() {
	op.initRequestProcessing()
}

// takes action when stateChanged flag is true
func (op *scOperator) setNewState(tx sc.Transaction) {
	op.stateTx = tx
	// reset current leader seq index
	op.currLeaderSeqIndex = 0
	op.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), op.stateTx.Id().Bytes())

	// swap arrays of incoming initReq's
	// clean the array of the next state
	op.requestToProcess[0], op.requestToProcess[1] = op.requestToProcess[1], op.requestToProcess[0]
	for i := range op.requestToProcess[1] {
		op.requestToProcess[1][i] = nil
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

func (op *scOperator) initRequestProcessing() {
	if !op.iAmCurrentLeader() {
		return
	}
	if op.currentRequest != nil {
		return
	}
	op.currentResult = nil
	op.currentRequest = op.selectRequestToProcess()
	if op.currentRequest == nil {
		return
	}
	msg := &initReqMsg{
		StateIndex: op.stateTx.MustState().StateIndex(),
		RequestId:  op.currentRequest.reqId,
	}
	var buf bytes.Buffer
	encodeInitReqMsg(msg, &buf)
	numSucc, ts := op.comm.SendMsgToPeers(msgInitRequest, buf.Bytes())

	if numSucc < op.Quorum() {
		op.currentRequest.log.Errorf("only %d 'msgInitRequest' sends succeeded", numSucc)
		op.currentRequest = nil
		return
	}
	op.currentRequest.log.Debugf("msgInitRequest successfully sent to %d peers", numSucc)
	op.asyncCalculateResult(op.currentRequest, ts)
}
