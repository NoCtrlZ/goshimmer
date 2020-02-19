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
		resState, _ := req.ownResultCalculated.res.state.State()
		resStateIndex := resState.StateIndex()
		curState, _ := op.stateTx.State()
		curStateIndex := curState.StateIndex()
		if resStateIndex != curStateIndex {
			req.ownResultCalculated = nil
			req.hasBeenPushedToCurrentLeader = false
		}
	}
	ktd1 := make([]*hashing.HashValue, 0)
	for pushMsg, lst := range req.receivedResultHashes {
		for _, rh := range lst {
			if rh != nil && !op.pushMsgConsistentWithContext(rh) {
				ktd1 = append(ktd1, pushMsg.Clone())
			}
		}
	}
	for _, k := range ktd1 {
		delete(req.receivedResultHashes, *k)
	}
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
		for _, rhlist := range req.receivedResultHashes {
			for _, rh := range rhlist {
				if rh != nil && !op.pushMsgConsistentWithContext(rh) {
					return fmt.Errorf("request hash out of context")
				}
			}
		}
		for _, am := range req.pullMessages {
			if !op.pullMsgConsistentWithContext(am) {
				return fmt.Errorf("pull message out of context")
			}
		}
	}
	return nil
}

func (op *AssemblyOperator) pushMsgConsistentWithContext(rh *pushResultMsg) bool {
	curState, _ := op.stateTx.State()
	return rh.StateIndex >= curState.StateIndex()
}

func (op *AssemblyOperator) pullMsgConsistentWithContext(am *pullResultMsg) bool {
	curState, _ := op.stateTx.State()
	return am.StateIndex >= curState.StateIndex()
}

func (op *AssemblyOperator) resultBelongsToContext(res *resultCalculated) bool {
	curState, _ := op.stateTx.State()
	resState, _ := res.state.State()
	return curState.StateIndex()+1 == resState.StateIndex()
}

func (op *AssemblyOperator) selectAdvanced() []*request {
	ret := make([]*request, 0)
	for _, req := range op.requests {
		if req.msgTx == nil {
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
	ret.Set("numProcessed", op.processedCounter)
	ret.Set("numPendingRequests", len(op.requests))

	for _, req := range op.requests {
		if req.ownResultCalculated != nil {
			ret.Inc("numNotResultCalculated")
		}
		numRhLst := getNumsResultHashes(req)
		if len(numRhLst) > 0 {
			ret.Inc("numWithResultHashes")
		}
		if len(numRhLst) > 1 {
			ret.Inc("numWithMoreThan1ResultHash")
		}
		if len(req.pullMessages) > 0 {
			ret.Inc("numWithPullMsgs")
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
	ret := make([]int, len(req.receivedResultHashes))
	i := 0
	for _, rhlst := range req.receivedResultHashes {
		for _, rh := range rhlst {
			if rh != nil {
				ret[i]++
			}
		}
		i++
	}
	return ret
}
