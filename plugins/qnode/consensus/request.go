package consensus

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"time"
)

// check if the request message is well formed
func (op *Operator) validateRequestBlock(reqRef *committee.RequestMsg) error {
	// TODO check rewards etc
	return nil
}

func (op *Operator) newRequest(reqId *sctransaction.RequestId) *request {
	reqLog := log.Named(reqId.Short())
	ret := &request{
		reqId: reqId,
		log:   reqLog,
	}
	reqLog.Info("NEW REQUEST")
	return ret
}

// request record retrieved (or created) by request message
func (op *Operator) requestFromMsg(reqMsg *committee.RequestMsg) *request {
	reqId := reqMsg.Id()
	ret, ok := op.requests[*reqId]
	if ok && ret.reqMsg == nil {
		ret.reqMsg = reqMsg
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqMsg = reqMsg
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret
}

// request record is retrieved by request id.
// If it doesn't exist and is not in the list of processed requests, it is created
func (op *Operator) requestFromId(reqId *sctransaction.RequestId) (*request, bool) {
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

func (op *Operator) isRequestProcessed(reqid *sctransaction.RequestId) (time.Duration, bool) {
	duration, ok := op.processedRequests[*reqid]
	return duration, ok
}

func (op *Operator) markRequestProcessed(req *request) {
	duration := time.Duration(0)
	if req.reqMsg != nil {
		duration = time.Since(req.whenMsgReceived)
	}
	req.log.Infof("REQUEST MARKED PROCESSED. duration since received: %v, msg count: %d",
		duration, req.msgCounter)
	op.processedRequests[*req.reqId] = duration
	delete(op.requests, *req.reqId)
}
