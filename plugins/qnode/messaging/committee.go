package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"sync"
)

var (
	committees     = make(map[hashing.HashValue]*CommitteeConn)
	committeeMutex = &sync.RWMutex{}
)

type CommitteeConn struct {
	operator         SCOperator
	recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte)
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

func RegisterNewOperator(op SCOperator, recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte)) *CommitteeConn {
	committeeMutex.Lock()
	defer committeeMutex.Unlock()

	if cconn, ok := committees[*op.SContractID()]; ok {
		return cconn
	}
	ret := &CommitteeConn{
		operator:         op,
		recvDataCallback: recvDataCallback,
		peers:            make([]*qnodePeer, len(op.NodeAddresses())),
	}
	for i := range ret.peers {
		if i == int(op.PeerIndex()) {
			continue
		}
		ret.peers[i] = addPeerConnection(op.NodeAddresses()[i])
	}
	committees[*op.SContractID()] = ret
	return ret
}

func (cconn *CommitteeConn) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == cconn.operator.PeerIndex() || int(targetPeerIndex) >= len(cconn.peers) {
		return fmt.Errorf("attempt to send message to wrong peer index")
	}
	wrapped := wrapPacket(cconn.operator.SContractID(), cconn.operator.PeerIndex(), msgType, msgData)
	return cconn.peers[targetPeerIndex].SendMsgData(wrapped)
}

//
//func (cconn *CommitteeConn) SendMsgToPeers(msgType byte, msgData []byte) uint16 {
//	wrapped := wrapPacket(cconn.operator.SContractID(), cconn.operator.PeerIndex(), msgType, msgData)
//	var sentTo uint16
//	for i, conn := range cconn.peers {
//		if i == int(cconn.operator.PeerIndex()) {
//			continue
//		}
//		if err := conn.SendMsgData(wrapped); err == nil {
//			log.Debugf("%v", err)
//			sentTo++
//		}
//	}
//	return sentTo
//}
