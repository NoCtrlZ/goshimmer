package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"sync"
	"time"
)

var (
	committees     = make(map[hashing.HashValue]*CommitteeConn)
	committeeMutex = &sync.RWMutex{}
)

type CommitteeConn struct {
	operator         SCOperator
	recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte, ts time.Time)
	peers            []*qnodePeer
}

func getCommittee(scid *hashing.HashValue) (*CommitteeConn, bool) {
	committeeMutex.RLock()
	defer committeeMutex.RUnlock()

	cconn, ok := committees[*scid]
	if !ok {
		return nil, false
	}
	return cconn, true
}

func GetOperator(scid *hashing.HashValue) (SCOperator, bool) {
	comm, ok := getCommittee(scid)
	if !ok {
		return nil, false
	}
	return comm.operator, true
}

func RegisterNewOperator(op SCOperator, recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte, ts time.Time)) *CommitteeConn {
	committeeMutex.Lock()
	defer committeeMutex.Unlock()

	if cconn, ok := committees[*op.SContractID()]; ok {
		return cconn
	}
	ret := &CommitteeConn{
		operator:         op,
		recvDataCallback: recvDataCallback,
		peers:            make([]*qnodePeer, len(op.PeerAddresses())),
	}
	for i := range ret.peers {
		if i == int(op.PeerIndex()) {
			continue
		}
		ret.peers[i] = addPeerConnection(op.PeerAddresses()[i])
	}
	committees[*op.SContractID()] = ret
	return ret
}

// sends message to specified peer
func (cconn *CommitteeConn) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == cconn.operator.PeerIndex() || int(targetPeerIndex) >= len(cconn.peers) {
		return fmt.Errorf("attempt to send message to the wrong peer index")
	}
	if msgType < FirstCommitteeMsgCode {
		panic("reserved msg type")
	}

	peer := cconn.peers[targetPeerIndex]

	var wrapped []byte

	wrapped, ts := wrapPacket(&unwrappedPacket{
		msgType:     msgType,
		scid:        cconn.operator.SContractID(),
		senderIndex: cconn.operator.PeerIndex(),
		data:        msgData,
	})

	peer.Lock()
	peer.lastHeartbeatSent = ts
	peer.Unlock()

	err := peer.sendData(wrapped)
	return err
}

// send message to peers.
// returns number if successful sends and timestamp common for all messages
func (cconn *CommitteeConn) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	if msgType == FirstCommitteeMsgCode {
		panic("reserved msg type")
	}

	var wrapped []byte
	wrapped, ts := wrapPacket(&unwrappedPacket{
		msgType:     msgType,
		scid:        cconn.operator.SContractID(),
		senderIndex: cconn.operator.PeerIndex(),
		data:        msgData,
	})
	var ret uint16

	for i := uint16(0); i < cconn.operator.CommitteeSize(); i++ {
		if i == cconn.operator.PeerIndex() {
			continue
		}
		peer := cconn.peers[i]
		peer.Lock()
		peer.lastHeartbeatSent = ts
		peer.Unlock()

		if err := peer.sendData(wrapped); err == nil {
			ret++
		}
	}
	return ret, ts
}
