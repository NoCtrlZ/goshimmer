package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
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

	// send until first success, but no more than number of nodes in the committee
	var i uint16
	for i = 0; i < op.CommitteeSize(); i++ {
		if op.iAmCurrentLeader() {
			// stop if I am the current leader
			return
		}
		if !op.comm.IsAlivePeer(op.currentLeaderPeerIndex()) {
			op.moveToNextLeader()
			continue
		}
		err := op.comm.SendMsg(op.currentLeaderPeerIndex(), msgNotifyRequests, buf.Bytes())
		if err == nil {
			op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
			// first node to which data was successfully sent is assumed the leader
			return
		}
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

func (op *scOperator) sortedRequestIdsByAge() []*sc.RequestId {
	sortedReqs := op.sortedRequestsByAge()
	ids := make([]*sc.RequestId, len(sortedReqs))
	for i := range ids {
		ids[i] = sortedReqs[i].reqId
	}
	return ids
}

// includes request ids into the respective list of notifications,
// by the sender index
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