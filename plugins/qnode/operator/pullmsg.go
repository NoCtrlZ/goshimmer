package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"time"
)

func (op *AssemblyOperator) sendPullMessages(res *resultCalculatedIntern, haveVotes uint16, maxVotedFor *hashing.HashValue) {
	reqId := RequestId(res.res.requestTx.Id(), res.res.requestIndex)
	state, _ := res.res.state.State()
	msg := &pullResultMsg{
		SenderIndex: op.peerIndex(),
		RequestId:   reqId,
		StateIndex:  state.StateIndex(),
		HaveVotes:   haveVotes,
	}
	reqRec, _ := op.requestFromIdHash(maxVotedFor)
	lst := reqRec.receivedResultHashes[*maxVotedFor]
	for idx, rh := range lst {
		if rh == nil && uint16(idx) != op.peerIndex() {
			err := op.sendMsgToPeer(msg, int16(idx))
			if err != nil {
				log.Errorf("SendUDPData returned error: `%v`", err)
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
		if req.msgTx == nil {
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
