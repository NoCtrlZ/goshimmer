// the committeecon package implements
// Conn object is responsible for the communications between committee nodes
package committeeconn

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/goshimmer/plugins/qnode/peering"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"time"
)

type Conn struct {
	ownIndex uint16
	scid     sctransaction.ScId
	peers    []*peering.Peer
}

func NewConnection(scid sctransaction.ScId, peers []*registry.PortAddr) (*Conn, error) {
	ownIndex, ok := peering.FindOwnIndex(peers)
	if !ok {
		return nil, fmt.Errorf("not processed by this node scid: %s", scid.String())
	}
	ret := &Conn{
		scid:     scid,
		ownIndex: ownIndex,
		peers:    make([]*peering.Peer, 0, len(peers))}
	for i, pa := range peers {
		if i != int(ownIndex) {
			ret.peers[i] = peering.UsePeer(pa)
		}
	}
	return ret, nil
}

// sends message to peer with index
func (conn *Conn) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == conn.ownIndex || int(targetPeerIndex) >= len(conn.peers) {
		return fmt.Errorf("SendMsg: wrong peer index")
	}
	peer := conn.peers[targetPeerIndex]
	msg := &events.PeerMessage{
		ScColor:     conn.scid.Color(),
		SenderIndex: conn.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (conn *Conn) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	msg := &events.PeerMessage{
		ScColor:     conn.scid.Color(),
		SenderIndex: conn.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, conn.peers...)
}

// returns true if peer is alive. Used by the operator to determine current leader
func (conn *Conn) IsAlivePeer(peerIndex uint16) bool {
	if peerIndex == conn.ownIndex {
		return true
	}
	if int(peerIndex) >= len(conn.peers) {
		return false
	}
	ret, _ := conn.peers[peerIndex].IsAlive()
	return ret
}
