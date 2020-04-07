package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// notifies current leader about requests in the order of arrival
func (op *scOperator) sendRequestNotificationsToLeader(reqs []*request) {
	ids := make([]*sc.RequestId, len(reqs))
	for i := range ids {
		ids[i] = reqs[i].reqId
	}
	msg := &notifyReqMsg{
		StateIndex: op.stateTx.MustState().StateIndex(),
		RequestIds: ids,
	}
	var buf bytes.Buffer
	encodeNotifyReqMsg(msg, &buf)

	err := op.comm.SendMsg(op.currentLeaderPeerIndex(), msgNotifyRequests, buf.Bytes())

	if err != nil {
		log.Errorf("sending req notifications: %v", err)
		return
	}
}

func (op *scOperator) sortedRequestsByAge() []*request {
	ret := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		ret = append(ret, req)
	}
	sortRequestsByAge(ret)
	return ret
}

func (op *scOperator) sendRequestNotificationsAllToLeader() {
	op.sendRequestNotificationsToLeader(op.sortedRequestsByAge())
}

func (op *scOperator) sendRequestNotification(req *request) {
	op.sendRequestNotificationsToLeader([]*request{req})
}

// includes request ids into the respective list of notifications
func (op *scOperator) accountRequestIdNotifications(senderIndex uint16, nextState bool, reqs ...*sc.RequestId) {
	pos := 0
	if nextState {
		pos = 1
	}
	for _, id := range reqs {
		op.requestNotificationsReceived[senderIndex][pos] =
			appendReqId(op.requestNotificationsReceived[senderIndex][pos], id)
	}
}

// ensures each id is unique in the list
func appendReqId(lst []*sc.RequestId, id *sc.RequestId) []*sc.RequestId {
	for _, tid := range lst {
		if tid.Equal(id) {
			return lst
		}
	}
	return append(lst, id)
}
