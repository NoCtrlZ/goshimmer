package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
)

// check if the request message is well formed
func (op *AssemblyOperator) validateRequestBlock(reqRef *sc.RequestRef) error {
	cfg := op.stateTx.MustState().Config()
	reward := uint64(0)
	rewardOutput := reqRef.RequestBlock().MainOutputs(reqRef.Tx()).RewardOutput
	if rewardOutput != nil {
		reward = rewardOutput.Value
	}
	if reward < cfg.MinimumReward() {
		return fmt.Errorf("reward is less than required minimum of %d", cfg.MinimumReward()+1)
	}
	return nil
}

func newRequest(reqId *HashValue) *request {
	reqLog := log.Named(reqId.Shortest())
	reqLog.Info("created new request record")
	return &request{
		reqId:              reqId,
		pushMessages:       make(map[HashValue][]*pushResultMsg),
		pullMessages:       make(map[uint16]*pullResultMsg),
		startedCalculation: make(map[HashValue]time.Time),
		log:                reqLog,
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

func (op *AssemblyOperator) markRequestProcessed(req *request, duration time.Duration) {
	req.log.Info("request marked 'processed'")
	op.processedRequests[*req.reqId] = duration
	delete(op.requests, *req.reqId)
}
