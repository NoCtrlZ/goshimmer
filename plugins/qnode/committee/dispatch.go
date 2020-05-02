package committee

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/commtypes"
	qnode_events "github.com/iotaledger/goshimmer/plugins/qnode/events"
	"time"
)

func (c *committee) dispatchMessage(msg interface{}) {
	if !c.isOperational.Load() {
		return
	}
	stateMgr := false

	switch msgt := msg.(type) {

	case *qnode_events.PeerMessage:
		// receive a message from peer
		c.processPeerMessage(msgt)

	case *commtypes.StateUpdateMsg:
		// StateUpdateMsg may come from peer and from own consensus operator
		c.stateMgr.EventStateUpdateMsg(msgt)

	case *commtypes.StateTransitionMsg:
		c.operator.EventStateTransitionMsg(msgt)

	case commtypes.StateTransactionMsg:
		// receive state transaction message
		c.stateMgr.EventStateTransactionMsg(msgt)

	case commtypes.RequestMsg:
		// receive request message
		c.operator.EventRequestMsg(msgt)

	case commtypes.TimerTick:
		if stateMgr {
			c.stateMgr.EventTimerMsg(msgt)
		} else {
			c.operator.EventTimerMsg(msgt)
		}
		stateMgr = !stateMgr
	}
}

func (c *committee) processPeerMessage(msg *qnode_events.PeerMessage) {
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
