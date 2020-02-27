package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"sort"
	"time"
)

// find existing request based on msg parameters or creates new one
// always returns request struct.
// returns flag if request message is new and other operators must be notified

func newRequest(reqId *HashValue) *request {
	return &request{
		reqId:              reqId,
		pushMessages:       make(map[HashValue][]*pushResultMsg),
		pullMessages:       make(map[uint16]*pullResultMsg),
		startedCalculation: make(map[HashValue]time.Time),
	}
}

// request record retrieved (or created) by request message

func (op *AssemblyOperator) requestFromMsg(reqRef *sc.RequestRef) *request {
	reqId := reqRef.Id()
	ret, ok := op.requests[*reqId]
	if ok && ret.reqRef == nil {
		ret.reqRef = reqRef
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqRef = reqRef
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret
}

// request record retrieved (or created) by request id

func (op *AssemblyOperator) requestFromId(reqId *HashValue) (*request, bool) {
	if _, yes := op.isRequestProcessed(reqId); yes {
		return nil, false
	}
	ret, ok := op.requests[*reqId]
	if !ok {
		ret = newRequest(reqId)
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret, true
}

func (op *AssemblyOperator) isRequestProcessed(reqid *HashValue) (time.Duration, bool) {
	duration, ok := op.processedRequests[*reqid]
	return duration, ok
}

func (op *AssemblyOperator) markRequestProcessed(reqId *HashValue, duration time.Duration) {
	sum := len(op.processedRequests) + len(op.requests)
	op.processedRequests[*reqId] = duration
	delete(op.requests, *reqId)
	sum1 := len(op.processedRequests) + len(op.requests)
	if sum != sum1 {
		log.Panicf("WTF!!!!!!")
	}
}

func maxVotesFromPeers(req *request) (uint16, *HashValue) {
	var retRsHash HashValue
	var retNumVotes uint16

	for rsHash, rhlst := range req.pushMessages {
		numNotNil := uint16(0)
		for _, rh := range rhlst {
			if rh != nil {
				numNotNil++
			}
		}
		if numNotNil > retNumVotes {
			retNumVotes = numNotNil
			copy(retRsHash.Bytes(), rsHash.Bytes())
		}
	}
	return retNumVotes, &retRsHash
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
