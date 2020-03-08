package operator

import (
	"bytes"
	"encoding/binary"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"sort"
	"time"
)

// sorting indices of operators by the request hash

type idxToPermute struct {
	idx  uint16
	hash *HashValue
}

type arrToSort []idxToPermute

func (s arrToSort) Len() int {
	return len(s)
}

func (s arrToSort) Less(i, j int) bool {
	return bytes.Compare((*s[i].hash)[:], (*s[j].hash)[:]) < 0
}

func (s arrToSort) Swap(i, j int) {
	s[i].idx, s[j].idx = s[j].idx, s[i].idx
	s[i].hash, s[j].hash = s[j].hash, s[i].hash
}

func getPermutation(n uint16, reqHash *HashValue) []uint16 {
	arr := make(arrToSort, n)
	var t [2]byte
	for i := range arr {
		binary.LittleEndian.PutUint16(t[:], uint16(i))
		arr[i] = idxToPermute{
			idx:  uint16(i),
			hash: HashData(reqHash.Bytes(), t[:]),
		}
	}
	sort.Sort(arr)
	ret := make([]uint16, n)
	for i := range ret {
		ret[i] = arr[i].idx
	}
	return ret
}

func (op *AssemblyOperator) iAmCurrentLeader(req *request) bool {
	return op.peerIndex() == op.currentLeaderIndex(req)
}

func (op *AssemblyOperator) currentLeaderIndex(req *request) uint16 {
	//return 3
	if req.leaderPeerIndexList == nil {
		req.leaderPeerIndexList = getPermutation(op.assemblySize(), req.reqId)
	}
	return req.leaderPeerIndexList[req.currLeaderSeqIndex]
}

func (op *AssemblyOperator) rotateLeaderIfNeeded(req *request) {
	if req.reqRef == nil || !req.hasBeenPushedToCurrentLeader {
		return
	}
	if time.Since(req.whenLastPushed) > parameters.LEADER_ROTATION_PERIOD {
		clead := req.currLeaderSeqIndex
		req.currLeaderSeqIndex = (req.currLeaderSeqIndex + 1) % int16(op.assemblySize())
		req.hasBeenPushedToCurrentLeader = false
		req.log.Infof("LEADER ROTATED %d --> %d", clead, req.currLeaderSeqIndex)
	}
}

func (op *AssemblyOperator) pushIfNeeded(req *request) {
	if !req.hasBeenPushedToCurrentLeader {
		op.sendPushResultToPeer(req.ownResultCalculated, op.currentLeaderIndex(req))
		req.hasBeenPushedToCurrentLeader = true
		req.whenLastPushed = time.Now()
	}
}
