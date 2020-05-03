package commiteeimpl

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	qnode_events "github.com/iotaledger/goshimmer/plugins/qnode/events"
	"time"
)

func (c *committeeObj) dispatchMessage(msg interface{}) {
	if !c.isOperational.Load() {
		return
	}
	stateMgr := false

	switch msgt := msg.(type) {

	case *qnode_events.PeerMessage:
		// receive a message from peer
		c.processPeerMessage(msgt)

	case *committee.StateUpdateMsg:
		// StateUpdateMsg may come from peer and from own consensus operator
		c.stateMgr.EventStateUpdateMsg(msgt)

	case *committee.StateTransitionMsg:
		c.operator.EventStateTransitionMsg(msgt)

	case committee.StateTransactionMsg:
		// receive state transaction message
		c.stateMgr.EventStateTransactionMsg(msgt)

	case committee.RequestMsg:
		// receive request message
		c.operator.EventRequestMsg(msgt)

	case committee.TimerTick:
		if stateMgr {
			c.stateMgr.EventTimerMsg(msgt)
		} else {
			c.operator.EventTimerMsg(msgt)
		}
		stateMgr = !stateMgr
	}
}

func (c *committeeObj) processPeerMessage(msg *qnode_events.PeerMessage) {
	rdr := bytes.NewReader(msg.MsgData)

	switch msg.MsgType {

	case committee.MsgNotifyRequests:
		msgt := &committee.NotifyReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex

		c.operator.EventNotifyReqMsg(msgt)

	case committee.MsgStartProcessingRequest:
		msgt := &committee.StartProcessingReqMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		c.operator.EventStartProcessingReqMsg(msgt)

	case committee.MsgSignedHash:
		msgt := &committee.SignedHashMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		msgt.Timestamp = time.Unix(0, msg.Timestamp)

		c.operator.EventSignedHashMsg(msgt)

	case committee.MsgGetStateUpdate:
		msgt := &committee.GetStateUpdateMsg{}
		if err := msgt.Read(rdr); err != nil {
			log.Error(err)
			return
		}
		msgt.SenderIndex = msg.SenderIndex
		c.stateMgr.EventGetStateUpdateMsg(msgt)

	case committee.MsgStateUpdate:
		msgt := &committee.StateUpdateMsg{}
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
