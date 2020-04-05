package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func (op *scOperator) iAmCurrentLeader(req *request) bool {
	return op.PeerIndex() == op.currentLeaderIndex(req)
}

func (op *scOperator) currentLeaderIndex(req *request) uint16 {
	//return 3
	if req.leaderPeerIndexList == nil {
		req.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), req.reqId.Bytes())
	}
	return req.leaderPeerIndexList[req.currLeaderSeqIndex]
}

func (op *scOperator) rotateLeaderIfNeeded(req *request) {
	if req.reqRef == nil || !req.hasBeenPushedToCurrentLeader {
		return
	}
	if time.Since(req.whenLastPushed) > parameters.LEADER_ROTATION_PERIOD {
		clead := req.currLeaderSeqIndex
		req.currLeaderSeqIndex = (req.currLeaderSeqIndex + 1) % int16(op.CommitteeSize())
		req.hasBeenPushedToCurrentLeader = false
		req.log.Infof("LEADER ROTATED %d --> %d", clead, req.currLeaderSeqIndex)
	}
}

func (op *scOperator) pushIfNeeded(req *request) {
	if !req.hasBeenPushedToCurrentLeader {
		op.sendPushResultToPeer(req.ownResultCalculated, op.currentLeaderIndex(req))
		req.hasBeenPushedToCurrentLeader = true
		req.whenLastPushed = time.Now()
	}
}
