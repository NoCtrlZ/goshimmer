package operator2

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

func (op *scOperator) iAmCurrentLeader() bool {
	return op.PeerIndex() == op.currentLeaderPeerIndex()
}

func (op *scOperator) currentLeaderPeerIndex() uint16 {
	return op.leaderAtSeqIndex(op.currLeaderSeqIndex)
}

func (op *scOperator) leaderAtSeqIndex(seqIdx uint16) uint16 {
	return op.leaderPeerIndexList[seqIdx]
}

const leaderRotationPeriod = 3 * time.Second

func (op *scOperator) moveToNextLeader() {
	op.currLeaderSeqIndex = (op.currLeaderSeqIndex + 1) % op.CommitteeSize()
	op.setLeaderRotationDeadline(time.Now().Add(leaderRotationPeriod))
}

func (op *scOperator) resetLeader() {
	op.currLeaderSeqIndex = 0
	op.leaderPeerIndexList = tools.GetPermutation(op.CommitteeSize(), op.stateTx.Id().Bytes())
	for i, v := range op.leaderPeerIndexList {
		if v == op.PeerIndex() {
			op.myLeaderSeqIndex = uint16(i)
			break
		}
	}
	// leader part of processing wasn't started yet
	op.leaderStatus = nil
	op.leaderRotationDeadlineSet = false
}

func (op *scOperator) setLeaderRotationDeadline(deadline time.Time) {
	op.leaderRotationDeadlineSet = false
	op.leaderRotationDeadline = deadline
}

func (op *scOperator) rotateLeaderIfNeeded() {
	if op.leaderRotationDeadlineSet && op.leaderRotationDeadline.After(time.Now()) {
		op.moveToNextLeader()
	}
}
