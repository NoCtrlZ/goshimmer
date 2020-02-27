package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func (op *AssemblyOperator) adjustToContext() {
	for _, req := range op.requests {
		op.adjustToContextReq(req)
	}
	if err := op.consistentState(); err != nil {
		log.Panicf("inconsistent stateTx after adjustToContext: %v", err)
	}
}

// delete all records from request which do not correspond to the new config id or stateTx id
func (op *AssemblyOperator) adjustToContextReq(req *request) {
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
	for pushMsg, lst := range req.pushMessages {
		for _, rh := range lst {
			if rh != nil && !op.pushMsgConsistentWithContext(rh) {
				ktd1 = append(ktd1, pushMsg.Clone())
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

func (op *AssemblyOperator) consistentState() error {
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

func (op *AssemblyOperator) pushMsgConsistentWithContext(rh *pushResultMsg) bool {
	return rh.StateIndex >= op.stateTx.MustState().StateIndex()
}

func (op *AssemblyOperator) pullMsgConsistentWithContext(am *pullResultMsg) bool {
	return am.StateIndex >= op.stateTx.MustState().StateIndex()
}

func (op *AssemblyOperator) resultBelongsToContext(res *runtimeContext) bool {
	// result didn't change during calculations
	return op.stateTx.MustState().StateIndex() == res.state.MustState().StateIndex()
}

// "advanced" are reqeuest record which:
//  - has recorded request message
//  - has some votes from peers

func (op *AssemblyOperator) selectAdvanced() []*request {
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

func (op *AssemblyOperator) getStateSnapshot() tools.StatMap {
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
			req.reqId.Short(), lst, op.currentLeaderIndex(req), result, pullSent, finalized)
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
