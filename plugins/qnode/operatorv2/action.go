package operator

import "bytes"

// each message is processed according to the same scheme:
// 1. adjust stateTx of the operator. stateTx must be consistent after this step
// 2. take action based on stateTx

func (op *scOperator) takeAction() {
	if err := op.consistentState(); err != nil {
		log.Errorf("inconsistent state: %v", err)
		return
	}
	op.sendRequestNotifications(false)
	op.initRequestProcessing()
}

func (op *scOperator) initRequestProcessing() {
	if !op.iAmCurrentLeader() {
		return
	}
	if op.currentRequest != nil {
		return
	}
	op.currentResult = nil
	op.currentRequest = op.selectRequestToProcess()
	if op.currentRequest == nil {
		return
	}
	msg := &initReqMsg{
		StateIndex: op.stateTx.MustState().StateIndex(),
		RequestId:  op.currentRequest.reqId,
	}
	var buf bytes.Buffer
	encodeInitReqMsg(msg, &buf)
	numSucc, ts := op.comm.SendMsgToPeers(msgInitRequest, buf.Bytes())

	if numSucc < op.Quorum() {
		op.currentRequest.log.Errorf("only %d 'msgInitRequest' sends succeeded", numSucc)
		op.currentRequest = nil
		return
	}
	op.currentRequest.log.Debugf("msgInitRequest successfully sent to %d peers", numSucc)
	op.asyncCalculateResult(op.currentRequest, ts)
}
