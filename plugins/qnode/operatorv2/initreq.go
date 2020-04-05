package operator

import "github.com/iotaledger/goshimmer/plugins/qnode/model/sc"

// analyzes notifications and selects request to process next
// return nil if request can't be selected
// this only makes sense if operator is the leader
func (op *scOperator) selectRequestToProcess() *request {
	stateIndex := op.stateTx.MustState().StateIndex()
	votes := make(map[sc.RequestId]int)
	for _, mapNotif := range op.requestNotificationsReceived {
		notifications, ok := mapNotif[stateIndex]
		if !ok {
			// don't cave notifications for the current state
			continue
		}
		for _, reqId := range notifications {
			if _, ok := votes[*reqId]; !ok {
				votes[*reqId] = 0
			}
			votes[*reqId] = votes[*reqId] + 1
		}
	}
	// calculate max votes
	maxVotes := 0
	for _, v := range votes {
		if v > maxVotes {
			maxVotes = v
		}
	}
	if uint16(maxVotes) < op.Quorum()-1 {
		// no request with at least quorum votes
		return nil
	}
	var ret *request
	for reqId, v := range votes {
		if v != maxVotes {
			continue
		}
		r := op.requests[reqId]
		if ret == nil || r.whenMsgReceived.Before(ret.whenMsgReceived) {
			ret = r
		}
	}
	return ret
}
