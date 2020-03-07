package operator

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"sort"
)

func (op *AssemblyOperator) validatePushMessage(msg *pushResultMsg) error {
	//if len(msg.SigBlocks) == 0 {
	//	return errors.New("push message with 0 blocks: invalid")
	//}
	return sc.VerifySignedBlocks(msg.SigBlocks, op)
}

func (op *AssemblyOperator) accountNewPushMsg(msg *pushResultMsg) error {
	if err := op.validatePushMessage(msg); err != nil {
		return err
	}
	resultHash := resultHash(msg.StateIndex, msg.RequestId, msg.MasterDataHash)
	req, _ := op.requestFromId(msg.RequestId)

	if rhlst, ok := req.pushMessages[*resultHash]; !ok {
		req.pushMessages[*resultHash] = make([]*pushResultMsg, op.assemblySize())
	} else {
		if rhlst[msg.SenderIndex] != nil {
			if !duplicatePushMessages(msg, rhlst[msg.SenderIndex]) {
				return fmt.Errorf("duplicate push msg has been received")
			}
			log.Warn("repeating push msg ignored")
		}
	}
	// if duplicate, replace the previous
	req.pushMessages[*resultHash][msg.SenderIndex] = msg
	return nil
}

func duplicatePushMessages(msg1, msg2 *pushResultMsg) bool {
	switch {
	case msg1 == msg2:
		return true
	case msg1.SenderIndex != msg2.SenderIndex:
		return false
	case !msg1.RequestId.Equal(msg2.RequestId):
		return false
	case msg1.StateIndex != msg2.StateIndex:
		return false
	case !msg1.MasterDataHash.Equal(msg2.MasterDataHash):
		return false
	case len(msg1.SigBlocks) != len(msg2.SigBlocks):
		return false
	default:
		for i := range msg1.SigBlocks {
			if !msg1.SigBlocks[i].SignedHash().Equal(msg2.SigBlocks[i].SignedHash()) {
				return false
			}
			if !msg1.SigBlocks[i].Account().Equal(msg2.SigBlocks[i].Account()) {
				return false
			}
		}
	}
	return true
}

func (op *AssemblyOperator) pickRequestToPush() *request {
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
