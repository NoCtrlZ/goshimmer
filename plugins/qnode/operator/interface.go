package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/pkg/errors"
	"time"
)

type SenderId struct {
	IpAddr string
	Port   int
	Index  uint16
}

func NewFromState(tx sc.Transaction, comm messaging.Messaging) (*AssemblyOperator, error) {
	return newFromState(tx, comm)
}

func (op *AssemblyOperator) ReceiveMsgData(sender SenderId, msgType byte, msgData []byte) error {
	if !op.validSender(sender) {
		return errors.New("invalid sender")
	}
	switch msgType {
	case MSG_PUSH_MSG:
		msg, err := decodePushResultMsg(sender.Index, msgData)
		if err != nil {
			return err
		}
		op.postEventToQueue(msg)

	case MSG_PULL_MSG:
		msg, err := decodePullResultMsg(sender.Index, msgData)
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

func (op *AssemblyOperator) startRoutines() {
	// start msg queue routine
	go func() {
		for msg := range op.inChan {
			op.dispatchEvent(msg)
		}
	}()
	// start clock tick routine
	if !parameters.TIMER_ON {
		return
	}
	chCancel := make(chan struct{})
	go func() {
		index := 0
		for {
			select {
			case <-chCancel:
				return
			case <-time.After(parameters.CLOCK_TICK_PERIOD):
				op.postEventToQueue(timerMsg(index))
				index++
			}
		}
	}()
	op.stopClock = func() {
		close(chCancel)
	}
}
