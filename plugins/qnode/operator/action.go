package operator

// each message is porcessed according to the same scheme:
// 1. adjust stateTx of the operator. stateTx must be consistent after this step
// 2. take action based on stateTx

func (op *AssemblyOperator) takeAction() {
	if err := op.consistentState(); err != nil {
		log.Errorf("inconsistent state: %v", err)
		return
	}
	op.rotateLeaders()
	op.sendPush()
	op.respondToPulls()
	op.checkQuorum()
	op.sendPull()
	op.startCalculations()
}

func (op *AssemblyOperator) rotateLeaders() {
	for _, req := range op.requests {
		op.rotateLeaderIfNeeded(req)
	}
}

func (op *AssemblyOperator) respondToPulls() {
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

func (op *AssemblyOperator) sendPull() {
	for _, req := range op.requests {
		if req.ownResultCalculated == nil || req.ownResultCalculated.pullSent {
			continue
		}
		if !op.iAmCurrentLeader(req) {
			continue
		}
		votes, votedHash := maxVotesFromPeers(req)
		if 0 < votes && votes < op.requiredQuorum()-1 {
			op.sendPullMessages(req.ownResultCalculated, votes, votedHash)
		}
	}
}

func (op *AssemblyOperator) checkQuorum() {
	for _, req := range op.requests {
		if req.ownResultCalculated == nil {
			continue
		}
		if req.ownResultCalculated.finalized {
			continue
		}
		maxVotes, maxVotedHash := maxVotesFromPeers(req)
		if !req.ownResultCalculated.rsHash.Equal(maxVotedHash) {
			// maybe voted from the future, the skip
			continue
		}
		if maxVotes+1 < op.requiredQuorum() {
			continue
		}
		// quorum reached for the current calculated result
		if err := op.finalizeTheRequest(req.ownResultCalculated); err != nil {
			log.Errorf("finalizeTheRequest: %v", err)
		}
		log.Debugf("finalized request %s", req.reqId.Short())
	}
}

func (op *AssemblyOperator) startCalculations() {
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

// one request to be processed
func (op *AssemblyOperator) sendPush() {
	// select best to sendPush (which is not led by me)
	req := op.pickRequestToPush()
	if req == nil {
		// nothing to process
		return
	}
	// assert req is not led by me
	if req.ownResultCalculated == nil {
		// start calculation no matter who is leading
		op.asyncCalculateResult(req)
		return
	}
	op.pushIfNeeded(req)
}
