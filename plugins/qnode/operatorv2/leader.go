package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func (op *scOperator) iAmCurrentLeader() bool {
	return op.PeerIndex() == op.currentLeaderIndex()
}

func (op *scOperator) currentLeaderIndex() uint16 {
	if op.leaderPeerIndexList == nil {
		op.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), op.stateTx.Id().Bytes())
	}
	return op.leaderPeerIndexList[op.currLeaderSeqIndex]
}

func (op *scOperator) rotateLeaderIfNeeded() {
	if !op.leaderRotationDeadlineSet || time.Now().Before(op.leaderRotationDeadline) {
		return
	}
	clead := op.currLeaderSeqIndex
	op.currLeaderSeqIndex = (op.currLeaderSeqIndex + 1) % int16(op.CommitteeSize())
	log.Infof("LEADER ROTATED %d --> %d", clead, op.currLeaderSeqIndex)

	req.hasBeenPushedToCurrentLeader = false

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
