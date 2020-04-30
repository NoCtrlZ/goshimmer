package committee

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/committeeconn"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/consensus"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/statemgr"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"time"
)

type committee struct {
	chMsg    chan interface{}
	scdata   *registry.SCData
	conn     *committeeconn.Conn
	stateMgr *statemgr.StateManager
	operator *consensus.Operator
}

func New(scdata *registry.SCData) (commtypes.Committee, error) {
	conn, err := committeeconn.NewConnection(scdata.ScId, scdata.NodeLocations)
	if err != nil {
		return nil, err
	}
	ret := &committee{
		chMsg:  make(chan interface{}, 10),
		scdata: scdata,
		conn:   conn,
	}
	ret.stateMgr = statemgr.NewStateManager(ret)
	ret.operator = consensus.NewOperator()

	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	return ret, nil
}

func (c *committee) ScId() sctransaction.ScId {
	return c.scdata.ScId
}

func (c *committee) Size() uint16 {
	return uint16(len(c.scdata.NodeLocations))
}

func (c *committee) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	return c.conn.SendMsg(targetPeerIndex, msgType, msgData)
}

func (c *committee) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	return c.conn.SendMsgToPeers(msgType, msgData)
}

func (c *committee) IsAlivePeer(peerIndex uint16) bool {
	return c.conn.IsAlivePeer(peerIndex)
}

func (c *committee) ProcessMessage(msg interface{}) {
	c.chMsg <- msg
}
