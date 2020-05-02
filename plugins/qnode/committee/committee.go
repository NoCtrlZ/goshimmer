package committee

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/consensus"
	"github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/peering"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/statemgr"
	"go.uber.org/atomic"
	"time"
)

type committee struct {
	isOperational atomic.Bool
	ownIndex      uint16
	peers         []*peering.Peer
	scdata        *registry.SCData
	chMsg         chan interface{}
	stateMgr      *statemgr.StateManager
	operator      *consensus.Operator
}

func New(scdata *registry.SCData) (commtypes.Committee, error) {
	ownIndex, ok := peering.FindOwnIndex(scdata.NodeLocations)
	if !ok {
		return nil, fmt.Errorf("not processed by this node scid: %s", scdata.ScId.String())
	}
	ret := &committee{
		ownIndex: ownIndex,
		chMsg:    make(chan interface{}, 10),
		scdata:   scdata,
		peers:    make([]*peering.Peer, 0, len(scdata.NodeLocations)),
	}
	for i, pa := range scdata.NodeLocations {
		if i != int(ownIndex) {
			ret.peers[i] = peering.UsePeer(pa)
		}
	}

	ret.stateMgr = statemgr.New(ret)
	ret.operator = consensus.NewOperator()

	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	if parameters.UseTimer {
		go func() {
			tick := 0
			for {
				time.Sleep(parameters.TimerTickPeriod)
				ret.ReceiveMessage(commtypes.TimerTick(tick))
			}
		}()
	}

	return ret, nil
}

// implements commtypes.Committee interface

func (c *committee) SetOperational() {
	c.isOperational.Store(true)
}

func (c *committee) Dismiss() {
	c.isOperational.Store(false)
	close(c.chMsg)
}

func (c *committee) ScId() sctransaction.ScId {
	return c.scdata.ScId
}

func (c *committee) Size() uint16 {
	return uint16(len(c.scdata.NodeLocations))
}

func (c *committee) ReceiveMessage(msg interface{}) {
	if c.isOperational.Load() {
		c.chMsg <- msg
	}
}

// sends message to peer with index
func (c *committee) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == c.ownIndex || int(targetPeerIndex) >= len(c.peers) {
		return fmt.Errorf("SendMsg: wrong peer index")
	}
	peer := c.peers[targetPeerIndex]
	msg := &events.PeerMessage{
		ScColor:     c.scdata.ScId.Color(),
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peer.SendMsg(msg)
}

func (c *committee) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	msg := &events.PeerMessage{
		ScColor:     c.scdata.ScId.Color(),
		SenderIndex: c.ownIndex,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	return peering.SendMsgToPeers(msg, c.peers...)
}

// returns true if peer is alive. Used by the operator to determine current leader
func (c *committee) IsAlivePeer(peerIndex uint16) bool {
	if peerIndex == c.ownIndex {
		return true
	}
	if int(peerIndex) >= len(c.peers) {
		return false
	}
	ret, _ := c.peers[peerIndex].IsAlive()
	return ret
}
