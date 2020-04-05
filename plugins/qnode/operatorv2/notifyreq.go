package operator

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// send request notifications to the current leader
// if renew == true sends all request ids
// if renew == false send only those which were not sent yet in the current context
func (op *scOperator) sendRequestNotifications(renew bool) {
	stateIndex := op.stateTx.MustState().StateIndex()
	reqsToSend := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		if renew {
			reqsToSend = append(reqsToSend, req)
		} else {
			if req.lastNotifiedLeaderOfStateIndex < stateIndex || req.lastNotifiedLeaderSeqIndex != op.currLeaderSeqIndex {
				reqsToSend = append(reqsToSend, req)
			}
		}
	}
	if len(reqsToSend) == 0 {
		return
	}
	sortRequestsByAge(reqsToSend)
	ids := make([]*sc.RequestId, len(reqsToSend))
	for i := range ids {
		ids[i] = reqsToSend[i].reqId
	}
	msg := &notifyReqMsg{
		StateIndex: stateIndex,
		Renew:      renew,
		RequestIds: ids,
	}
	var buf bytes.Buffer
	encodeNotifyReqMsg(msg, &buf)

	err := op.comm.SendMsg(op.currentLeaderPeerIndex(), msgNotifyRequests, buf.Bytes())

	if err != nil {
		log.Errorf("sending req notifications: %v", err)
		return
	}
	for _, req := range reqsToSend {
		req.lastNotifiedLeaderOfStateIndex = stateIndex
		req.lastNotifiedLeaderSeqIndex = op.currLeaderSeqIndex
	}
}
