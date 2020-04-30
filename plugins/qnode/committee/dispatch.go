package committee

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	qnode_events "github.com/iotaledger/goshimmer/plugins/qnode/events"
	"time"
)

func (c *Committee) ProcessMessage(msg interface{}) {
	c.chMsg <- msg
}

func (c *Committee) dispatchMessage(msg interface{}) {
	if c.stateMgr.IsCorruptedState() {
		return
	}
	switch msgt := msg.(type) {

	case *qnode_events.PeerMessage:
		c.processPeerMessage(msgt)

	case commtypes.StateTransactionMsg:
		c.stateMgr.EventStateTransactionMsg(msgt)

	case commtypes.RequestMsg:
		c.operator.EventRequestMsg(msgt)
	}
}

func (c *Committee) processPeerMessage(msg *qnode_events.PeerMessage) {
	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case commtypes.MsgNotifyRequests:
		msgt := &commtypes.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		c.operator.EventNotifyReqMsg(msgt)

	case commtypes.MsgStartProcessingRequest:
		msgt := &commtypes.StartProcessingReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		c.operator.EventStartProcessingReqMsg(msgt)

	case commtypes.MsgSignedHash:
		msgt := &commtypes.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		c.operator.EventSignedHashMsg(msgt)

	case commtypes.MsgGetStateUpdate:
		msgt := &commtypes.GetStateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventGetStateUpdateMsg(msgt)

	case commtypes.MsgStateUpdate:
		msgt := &commtypes.StateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventStateUpdateMsg(msgt)

	default:
		log.Errorf("processPeerMessage: wrong msg type")
	}
}
