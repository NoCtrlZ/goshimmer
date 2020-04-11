package operator2

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"time"
)

// check if the request message is well formed
func (op *scOperator) validateRequestBlock(reqRef *sc.RequestRef) error {
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

func (op *scOperator) newRequest(reqId *sc.RequestId) *request {
	reqLog := log.Named(reqId.Shortest())
	ret := &request{
		reqId: reqId,
		log:   reqLog,
	}
	reqLog.Info("NEW REQUEST")
	return ret
}

// request record retrieved (or created) by request message
func (op *scOperator) requestFromMsg(reqRef *sc.RequestRef) *request {
	reqId := reqRef.Id()
	ret, ok := op.requests[*reqId]
	if ok && ret.reqRef == nil {
		ret.reqRef = reqRef
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqRef = reqRef
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret
}

// request record is retrieved by request id.
// If it doesn't exist and is not in the list of processed requests, it is created
func (op *scOperator) requestFromId(reqId *sc.RequestId) (*request, bool) {
	if _, yes := op.isRequestProcessed(reqId); yes {
		return nil, false
	}
	ret, ok := op.requests[*reqId]
	if !ok {
		ret = op.newRequest(reqId)
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret, true
}

func (op *scOperator) isRequestProcessed(reqid *sc.RequestId) (time.Duration, bool) {
	duration, ok := op.processedRequests[*reqid]
	return duration, ok
}

func (op *scOperator) markRequestProcessed(req *request) {
	duration := time.Duration(0)
	if req.reqRef != nil {
		duration = time.Since(req.whenMsgReceived)
	}
	req.log.Infof("REQUEST MARKED PROCESSED. duration since received: %v, msg count: %d",
		duration, req.msgCounter)
	op.processedRequests[*req.reqId] = duration
	delete(op.requests, *req.reqId)
}
