package operator

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"time"
)

func (op *AssemblyOperator) sendPullMessages(res *resultCalculated, haveVotes uint16, maxVotedForResultHash *hashing.HashValue) {
	reqId := res.res.reqRef.Id()
	state, _ := res.res.state.State()
	msg := &pullResultMsg{
		SenderIndex: op.PeerIndex(),
		RequestId:   reqId,
		StateIndex:  state.StateIndex(),
		HaveVotes:   haveVotes,
	}
	reqRec, ok := op.requestFromId(reqId)
	if !ok {
		log.Panic("inconsistency: no request found")
	}
	lst := reqRec.pushMessages[*maxVotedForResultHash]
	for idx, rh := range lst {
		if rh == nil && uint16(idx) != op.PeerIndex() {
			var buf bytes.Buffer
			encodePullResultMsg(msg, &buf)
			err := op.comm.SendMsg(uint16(idx), MSG_PULL_MSG, buf.Bytes())
			if err != nil {
				log.Error(err)
			}
		}
	}
	res.pullSent = true
	res.whenLastPullSent = time.Now()
}

func pullMsgMaxVotes(req *request) (uint16, uint16) {
	var maxHaveVotes uint16
	var retPeer uint16
	for peer, am := range req.pullMessages {
		if am.HaveVotes > maxHaveVotes {
			maxHaveVotes = am.HaveVotes
			retPeer = uint16(peer)
		}
	}
	return maxHaveVotes, retPeer
}

func (op *AssemblyOperator) selectRequestToRespondToPullMsg() (*request, uint16) {
	var ret *request
	var retPeer uint16
	var maxVotes uint16
	for _, req := range op.requests {
		if req.reqRef == nil {
			continue
		}
		if len(req.pullMessages) == 0 {
			continue
		}
		v, peer := pullMsgMaxVotes(req)
		if v > maxVotes {
			v = maxVotes
			ret = req
			retPeer = peer
		}
	}
	return ret, retPeer
}
