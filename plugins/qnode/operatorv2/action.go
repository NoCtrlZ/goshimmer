package operator

import "bytes"

// each message is porcessed according to the same scheme:
// 1. adjust stateTx of the operator. stateTx must be consistent after this step
// 2. take action based on stateTx

func (op *scOperator) takeAction() {
	if err := op.consistentState(); err != nil {
		log.Errorf("inconsistent state: %v", err)
		return
	}
	op.sendRequestNotifications(false)
	op.rotateLeaders()
	op.initRequest()
	op.respondToPulls()
	op.checkQuorum()
	op.sendPull()
	op.startCalculations()
}

func (op *scOperator) rotateLeaders() {
	for _, req := range op.requests {
		op.rotateLeaderIfNeeded(req)
	}
}

func (op *scOperator) respondToPulls() {
	req, peer := op.selectRequestToRespondToPullMsg()
	if req == nil {
		return
	}
	if req.ownResultCalculated == nil {
		op.asyncCalculateResult(req)
		return
	}
	op.sendPushResultToPeer(req.ownResultCalculated, peer)
	delete(req.pullMessages, peer)
	//tools.Logf(1, "responded to pull from peer %d for req %s", peer, req.reqId.ShortStr())
}

func (op *scOperator) sendPull() {
	for _, req := range op.requests {
		if req.ownResultCalculated == nil || req.ownResultCalculated.pullSent {
			continue
		}
		if !op.iAmCurrentLeader(req) {
			continue
		}
		votes, votedHash := maxVotesFromPeers(req)
		if 0 < votes && votes < op.Quorum()-1 {
			op.sendPullMessages(req.ownResultCalculated, votes, votedHash)
		}
	}
}

func (op *scOperator) checkQuorum() {
	for _, req := range op.requests {
		if req.ownResultCalculated == nil {
			continue
		}
		if req.ownResultCalculated.finalized {
			continue
		}
		maxVotes, maxVotedHash := maxVotesFromPeers(req)
		if !req.ownResultCalculated.resultHash.Equal(maxVotedHash) {
			// maybe voted from the future, then skip
			continue
		}
		if maxVotes+1 < op.Quorum() {
			continue
		}
		// quorum reached for the current calculated result
		op.finalizeTheRequest(req.ownResultCalculated)
		log.Debugf("finalized request %s", req.reqId.Short())
	}
}

func (op *scOperator) startCalculations() {
	for _, req := range op.requests {
		if req.reqRef == nil && req.ownResultCalculated != nil {
			continue
		}
		votes, _ := maxVotesFromPeers(req)
		if votes == 0 && len(req.pullMessages) == 0 {
			continue
		}
		op.asyncCalculateResult(req)
	}
}

func (op *scOperator) initRequest() {
	if !op.iAmCurrentLeader() {
		return
	}
	if op.currentRequest != nil {
		return
	}
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
