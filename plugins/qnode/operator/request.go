package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"time"
)

// check if the request message is well formed
func (op *AssemblyOperator) validateRequest(reqRef *sc.RequestRef) error {
	cfg := op.stateTx.MustState().Config()
	sum := value.SumOutputsToAddress(reqRef.Tx().Transfer(), cfg.AssemblyAccount())
	if sum < cfg.MinimumReward()+1 {
		return fmt.Errorf("reward %d iotas is less than required minimum of %d", sum, cfg.MinimumReward()+1)
	}
	return nil
}

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
