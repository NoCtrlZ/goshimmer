package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
)

// notifies current leader about requests in the order of arrival
func (op *scOperator) sendRequestNotificationsToLeader(reqs []*request) {
	if op.iAmCurrentLeader() {
		return
	}
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

// only requests with reqRef != nil
func (op *scOperator) sortedRequestsByAge() []*request {
	ret := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		if req.reqRef != nil {
			ret = append(ret, req)
		}
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
func (op *scOperator) accountRequestIdNotifications(senderIndex uint16, stateIndex uint32, reqs ...*sc.RequestId) {

	switch {
	case stateIndex == op.stateTx.MustState().StateIndex():
		for _, id := range reqs {
			op.requestNotificationsCurrentState = appendNotification(op.requestNotificationsCurrentState, id, senderIndex)
		}
	case stateIndex == op.stateTx.MustState().StateIndex()+1:
		for _, id := range reqs {
			op.requestNotificationsNextState = appendNotification(op.requestNotificationsNextState, id, senderIndex)
		}
	}
}

// ensures each id is unique in the list
func appendNotification(lst []*requestNotification, id *sc.RequestId, peerIndex uint16) []*requestNotification {
	for _, tid := range lst {
		if tid.reqId.Equal(id) && tid.peerIndex == peerIndex {
			return lst
		}
	}
	return append(lst, &requestNotification{
		reqId:     id,
		peerIndex: peerIndex,
	})
}
