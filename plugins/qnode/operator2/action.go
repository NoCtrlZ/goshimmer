package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func (op *scOperator) takeAction() {
	op.adjustToStateContext()
	op.sendRequestNotifications(false)
	op.initRequestProcessing()
}

// takes action when stateChanged flag is true
func (op *scOperator) adjustToStateContext() {
	if !op.stateChanged {
		return
	}
	// reset current leader seq index
	op.currLeaderSeqIndex = 0
	op.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), op.stateTx.Id().Bytes())

	// swap arrays of incoming initReq's
	// clean the array of the next state
	op.requestToProcess[0], op.requestToProcess[1] = op.requestToProcess[1], op.requestToProcess[0]
	for i := range op.requestToProcess[1] {
		op.requestToProcess[1][i] = nil
	}
	// send request notification to peers with renew flag = true
	op.adjustNotifications()
	op.sendRequestNotifications(true)
	op.stateChanged = false
}

// delete notifications which belongs to the past state indices
func (op *scOperator) adjustNotifications() {
	obsoleteIndices := make([]uint32, 0)
	currentStateIndex := op.stateTx.MustState().StateIndex()
	for _, notifMap := range op.requestNotificationsReceived {
		for stateIndex := range notifMap {
			if stateIndex < currentStateIndex {
				obsoleteIndices = append(obsoleteIndices, stateIndex)
			}
		}
		for _, idx := range obsoleteIndices {
			delete(notifMap, idx)
		}
		obsoleteIndices = obsoleteIndices[:0]
	}
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
