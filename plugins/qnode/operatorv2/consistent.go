package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func (op *scOperator) adjustToContext() {
	op.adjustNotifications()
	for _, req := range op.requests {
		op.adjustToContextReq(req)
	}
	if err := op.consistentState(); err != nil {
		log.Panicf("inconsistent stateTx after adjustToContext: %v", err)
	}
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

func (op *scOperator) checkForCheating(req *request) {
	// simplified: only checking for cheating when own request is known
	// and it belongs to context
	if req.ownResultCalculated == nil {
		return
	}
	if !op.resultBelongsToContext(req.ownResultCalculated.res) {
		return
	}
	if len(req.pushMessages) <= 1 {
		return
	}
	for _, rhlst := range req.pushMessages {
		for _, msg := range rhlst {
			if msg == nil {
				continue
			}
			if msg.StateIndex != op.stateTx.MustState().StateIndex() {
				continue // possibly from the future
			}
			if !msg.MasterDataHash.Equal(req.ownResultCalculated.masterDataHash) {
				req.log.Warnf("!!! unexpected master hash from peer's %d: someone is cheating", msg.SenderIndex)
			}
		}
	}
}

// delete all records from request which do not correspond to the new config id or stateTx id
func (op *scOperator) adjustToContextReq(req *request) {
	if req.ownResultCalculated != nil {
		resStateIndex := req.ownResultCalculated.res.state.MustState().StateIndex()
		curStateIndex := op.stateTx.MustState().StateIndex()
		if resStateIndex != curStateIndex {
			req.ownResultCalculated = nil
			req.hasBeenPushedToCurrentLeader = false
		}
	}
	// delete push messages which are not from the current context
	ktd1 := make([]*hashing.HashValue, 0)
	for resultHash, lst := range req.pushMessages {
		for _, rh := range lst {
			if rh != nil && !op.pushMsgConsistentWithContext(rh) {
				ktd1 = append(ktd1, resultHash.Clone())
				break
			}
		}
	}
	for _, k := range ktd1 {
		delete(req.pushMessages, *k)
	}
	// delete pull messages which are not from the current context
	ktd2 := make([]uint16, 0)
	for k, v := range req.pullMessages {
		if !op.pullMsgConsistentWithContext(v) {
			ktd2 = append(ktd2, k)
		}
	}
	for _, k := range ktd2 {
		delete(req.pullMessages, k)
	}
}

func (op *scOperator) consistentState() error {
	for _, req := range op.requests {
		if req.ownResultCalculated != nil && !op.resultBelongsToContext(req.ownResultCalculated.res) {
			return fmt.Errorf("request result out of context")
		}
		for _, rhlist := range req.pushMessages {
			for _, rh := range rhlist {
				if rh != nil && !op.pushMsgConsistentWithContext(rh) {
					return fmt.Errorf("push msg out of context")
				}
			}
		}
		for _, am := range req.pullMessages {
			if !op.pullMsgConsistentWithContext(am) {
				return fmt.Errorf("pull msg out of context")
			}
		}
	}
	return nil
}

func (op *scOperator) pushMsgConsistentWithContext(rh *pushResultMsg) bool {
	return rh.StateIndex >= op.stateTx.MustState().StateIndex()
}

func (op *scOperator) pullMsgConsistentWithContext(am *pullResultMsg) bool {
	return am.StateIndex >= op.stateTx.MustState().StateIndex()
}

func (op *scOperator) resultBelongsToContext(res *runtimeContext) bool {
	// result didn't change during calculations
	return op.stateTx.MustState().StateIndex() == res.state.MustState().StateIndex()
}

// "advanced" are reqeuest record which:
//  - has recorded request message
//  - has some votes from peers

func (op *scOperator) selectAdvanced() []*request {
	ret := make([]*request, 0)
	for _, req := range op.requests {
		if req.reqRef == nil {
			// only those with request message received
			continue
		}
		if maxNumVotes, _ := maxVotesFromPeers(req); maxNumVotes == 0 {
			// only those which has at least one result hash received
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

func (op *scOperator) getStateSnapshot() tools.StatMap {
	ret := make(tools.StatMap)
	ret.Set("msgCounter", op.msgCounter)
	ret.Set("numRequestsProcessed", len(op.processedRequests))
	ret.Set("numPendingRequests", len(op.requests))

	if len(op.processedRequests) == 11 && len(op.requests) > 0 {
		log.Debugw("kuku")
	}

	for _, req := range op.requests {
		if req.ownResultCalculated != nil {
			ret.Inc("numResultCalculated")
		}
		numRhLst := getNumsResultHashes(req)
		if len(numRhLst) > 0 {
			ret.Inc("numWithResultHashes")
		}
	}
	advanced := op.selectAdvanced()
	maHashes := make([]string, 0)
	for _, req := range advanced {
		lst := getNumsResultHashes(req)
		result := req.ownResultCalculated != nil
		finalized := result && req.ownResultCalculated.finalized
		pullSent := ""
		if result && req.ownResultCalculated.pullSent {
			pullSent = fmt.Sprintf("%v ago", time.Since(req.ownResultCalculated.whenLastPullSent))
		} else {
			pullSent = "no"
		}
		s := fmt.Sprintf("%s: %+v: lead=%d res=%v pullSent=%s final=%v",
			req.reqId.Short(), lst, op.currentLeaderPeerIndex(req), result, pullSent, finalized)
		maHashes = append(maHashes, s)
	}
	ret.Set("advanced", maHashes)
	ret.Set("advanced_total", len(advanced))
	for _, req := range advanced {
		if req.ownResultCalculated != nil {
			ret.Inc("advanced_withResult")
			if req.ownResultCalculated.finalized {
				ret.Inc("advanced_withFinalizedResult")
			}
			if req.hasBeenPushedToCurrentLeader {
				ret.MinInt64("advanced_lastPushedToLeader_msecAgo", time.Since(req.whenLastPushed).Milliseconds())
			}
		}
	}
	return ret
}

func getNumsResultHashes(req *request) []int {
	ret := make([]int, len(req.pushMessages))
	i := 0
	for _, rhlst := range req.pushMessages {
		for _, rh := range rhlst {
			if rh != nil {
				ret[i]++
			}
		}
		i++
	}
	return ret
}
