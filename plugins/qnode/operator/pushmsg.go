package operator

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"sort"
)

func (op *scOperator) accountNewPushMsg(msg *pushResultMsg) {
	resultHash := resultHash(msg.StateIndex, msg.RequestId, msg.MasterDataHash)
	req, ok := op.requestFromId(msg.RequestId)
	if !ok {
		// request is already processed
		return
	}
	if err := sc.VerifySignedBlocks(msg.SigBlocks, op.keyPool()); err != nil {
		req.log.Errorf("accountNewPushMsg: %v", err)
		return
	}

	if _, ok := req.pushMessages[*resultHash]; !ok {
		req.log.Debugf("new result hash %s from %d", resultHash.Short(), msg.SenderIndex)
		req.pushMessages[*resultHash] = make([]*pushResultMsg, op.CommitteeSize())
	}
	if req.pushMessages[*resultHash][msg.SenderIndex] != nil {
		// already received result hash from the same peer and the same result hash
		req.log.Warn("accountNewPushMsg: repeating push msg ignored")
	} else {
		req.pushMessages[*resultHash][msg.SenderIndex] = msg
	}
	req.log.Debugf("number of received push messages: %d", numberPushMessagesReceived(req.pushMessages[*resultHash]))
	op.checkForCheating(req)
}

func numberPushMessagesReceived(msgs []*pushResultMsg) uint16 {
	var ret uint16
	for _, m := range msgs {
		if m != nil {
			ret++
		}
	}
	return ret
}

func (op *scOperator) pickRequestToPush() *request {
	// with request message received and not led by me
	reqs := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		if req.reqRef == nil {
			continue
		}
		if op.iAmCurrentLeader(req) {
			continue
		}
		if req.hasBeenPushedToCurrentLeader {
			// only one is pushed each moment
			return nil
		}
		reqs = append(reqs, req)
	}
	if len(reqs) == 0 {
		return nil
	}
	// select the oldest 5
	sortRequestsByAge(reqs)
	if len(reqs) > parameters.NUM_OLDEST_REQESTS {
		reqs = reqs[:parameters.NUM_OLDEST_REQESTS]
	}
	// select the one with minimal id
	sortRequestsById(reqs)
	return reqs[0]
}

type sortByAge []*request

func (s sortByAge) Len() int {
	return len(s)
}

func (s sortByAge) Less(i, j int) bool {
	return s[i].whenMsgReceived.Before(s[j].whenMsgReceived)
}

func (s sortByAge) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortRequestsByAge(reqs []*request) {
	sort.Sort(sortByAge(reqs))
}

type sortById []*request

func (s sortById) Len() int {
	return len(s)
}

func (s sortById) Less(i, j int) bool {
	return bytes.Compare(s[i].reqId.Bytes(), s[j].reqId.Bytes()) < 0
}

func (s sortById) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortRequestsById(reqs []*request) {
	sort.Sort(sortById(reqs))
}
