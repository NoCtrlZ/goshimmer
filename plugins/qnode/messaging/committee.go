package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

type CommitteeConn struct {
	operator    SCOperator
	connections []*qnodePeer
}

func GetOperator(scid *hashing.HashValue) (SCOperator, bool) {
	peersMutex.RLock()
	defer peersMutex.RUnlock()

	cconn, ok := committees[*scid]
	if !ok {
		return nil, false
	}
	return cconn.operator, true
}

func RegisterNewOperator(op SCOperator) *CommitteeConn {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	if cconn, ok := committees[*op.SContractID()]; ok {
		return cconn
	}
	ret := &CommitteeConn{
		operator:    op,
		connections: make([]*qnodePeer, len(op.NodeAddresses())),
	}
	for i := range ret.connections {
		if i == int(op.PeerIndex()) {
			continue
		}
		ret.connections[i] = addPeerConnection_(op.NodeAddresses()[i])
	}
	committees[*op.SContractID()] = ret
	return ret
}

func (cconn *CommitteeConn) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == cconn.operator.PeerIndex() || int(targetPeerIndex) >= len(cconn.connections) {
		return fmt.Errorf("attempt to send message to wrong peer index")
	}
	wrapped := wrapPacket(cconn.operator.SContractID(), cconn.operator.PeerIndex(), msgType, msgData)
	return cconn.connections[targetPeerIndex].sendMsgData(wrapped)
}

//
//func (cconn *CommitteeConn) SendMsgToPeers(msgType byte, msgData []byte) uint16 {
//	wrapped := wrapPacket(cconn.operator.SContractID(), cconn.operator.PeerIndex(), msgType, msgData)
//	var sentTo uint16
//	for i, conn := range cconn.connections {
//		if i == int(cconn.operator.PeerIndex()) {
//			continue
//		}
//		if err := conn.sendMsgData(wrapped); err == nil {
//			log.Debugf("%v", err)
//			sentTo++
//		}
//	}
//	return sentTo
//}
