package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"sort"
	"time"
)

// find existing request based on msg parameters or creates new one
// always returns request struct.
// returns flag if request message is new and other operators must be notified

func newRequest(reqId *HashValue) *request {
	return &request{
		reqId:                reqId,
		receivedResultHashes: make(map[HashValue][]*pushResultMsg),
		startedCalculation:   make(map[HashValue]time.Time),
		pullMessages:         make(map[uint16]*pullResultMsg),
	}
}

// request record retrieved (or created) by request message

func (op *AssemblyOperator) requestFromMsg(tx sc.Transaction, reqIndex uint16) *request {
	reqId := RequestIdFromTx(tx, reqIndex)
	ret, ok := op.requests[*reqId]
	if ok && ret.msgTx == nil {
		ret.msgTx = tx
		ret.msgIndex = reqIndex
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.msgTx = tx
		ret.msgIndex = reqIndex
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret
}

func RequestIdFromTx(tx sc.Transaction, reqIndex uint16) *HashValue {
	return RequestId(tx.Id(), reqIndex)
}

func RequestId(txhash *HashValue, reqIndex uint16) *HashValue {
	return HashData(txhash.Bytes(), tools.Uint16To2Bytes(reqIndex))
}

// request record retrieved (or created) by request id

func (op *AssemblyOperator) requestFromIdHash(reqIdHash *HashValue) (*request, bool) {
	created := false
	ret, ok := op.requests[*reqIdHash]
	if !ok {
		ret = newRequest(reqIdHash)
		op.requests[*reqIdHash] = ret
		created = true
	}
	ret.msgCounter++
	return ret, created
}

func maxVotesFromPeers(req *request) (uint16, *HashValue) {
	var retResHash HashValue
	var retNumVotes uint16

	for resHash, rhlst := range req.receivedResultHashes {
		numNotNil := uint16(0)
		for _, rh := range rhlst {
			if rh != nil {
				numNotNil++
			}
		}
		if numNotNil > retNumVotes {
			retNumVotes = numNotNil
			copy(retResHash.Bytes(), resHash.Bytes())
		}
	}
	return retNumVotes, &retResHash
}

// TODO to config params
const numOldest = 5

func (op *AssemblyOperator) pickRequestToPush() *request {
	// with request message received and not led by me
	reqs := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		if req.msgTx == nil {
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
	if len(reqs) > numOldest {
		reqs = reqs[:numOldest]
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
