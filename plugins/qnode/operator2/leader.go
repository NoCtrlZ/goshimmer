package operator2

func (op *scOperator) iAmCurrentLeader() bool {
	return op.PeerIndex() == op.currentLeaderPeerIndex()
}

func (op *scOperator) currentLeaderPeerIndex() uint16 {
	return op.leaderPeerIndexList[op.currLeaderSeqIndex]
}
