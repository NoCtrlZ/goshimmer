package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/pkg/errors"
	"net"
)

func (op *AssemblyOperator) ReceiveUDPData(udpAddr *net.UDPAddr, senderIndex uint16, msgType byte, msgData []byte) error {
	if !op.validSender(udpAddr, senderIndex) {
		return errors.New("invalid sender")
	}
	switch msgType {
	case MSG_PUSH_MSG:
		msg, err := decodePushResultMsg(senderIndex, msgData)
		if err != nil {
			return err
		}
		op.postEventToQueue(msg)

	case MSG_PULL_MSG:
		msg, err := decodePullResultMsg(senderIndex, msgData)
		if err != nil {
			return err
		}
		op.postEventToQueue(msg)

	default:
		return errors.New("wrong msg type")
	}
	return nil
}

func (op *AssemblyOperator) ReceiveStateUpdate(msg *sc.StateUpdateMsg) {
	op.postEventToQueue(msg)
}

func (op *AssemblyOperator) ReceiveRequest(msg *sc.RequestRef) {
	op.postEventToQueue(msg)
}

func (op *AssemblyOperator) IsDismissed() bool {
	op.RLock()
	defer op.RUnlock()
	return op.dismissed
}

func (op *AssemblyOperator) dismiss() {
	if op.stopClock != nil {
		op.stopClock()
	}
	op.Lock()
	op.dismissed = true
	close(op.inChan)
	op.Unlock()
}

func (op *AssemblyOperator) postEventToQueue(msg interface{}) {
	op.RLock()
	defer op.RUnlock()
	if !op.dismissed {
		op.inChan <- msg
	}
}

func (op *AssemblyOperator) dispatchEvent(msg interface{}) {
	if _, ok := msg.(timerMsg); !ok {
		op.msgCounter++
	}
	switch msgt := msg.(type) {
	case *sc.StateUpdateMsg:
		op.eventStateUpdate(msgt.Tx)
	case *sc.RequestRef:
		op.eventRequestMsg(msgt)
	case *runtimeContext:
		op.eventResultCalculated(msgt)
	case *pushResultMsg:
		op.eventPushResultMsg(msgt)
	case *pullResultMsg:
		op.eventPullMsgReceived(msgt)
	case timerMsg:
		op.eventTimer(msgt)
	default:
		log.Panicf("dispatchEvent: wrong message type %T", msg)
	}
}
