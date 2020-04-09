package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"time"
)

func (op *scOperator) SContractID() *hashing.HashValue {
	return op.scid
}

func (op *scOperator) Quorum() uint16 {
	return op.cfgData.T
}

func (op *scOperator) CommitteeSize() uint16 {
	return op.cfgData.N
}

func (op *scOperator) PeerIndex() uint16 {
	return op.cfgData.Index
}

func (op *scOperator) PeerAddresses() []*registry.PortAddr {
	return op.cfgData.NodeAddresses
}

func NewFromState(tx sc.Transaction) (*scOperator, error) {
	return newFromState(tx)
}

func (op *scOperator) ReceiveStateUpdate(msg *sc.StateUpdateMsg) {
	op.postEventToQueue(msg)
}

func (op *scOperator) ReceiveRequest(msg *sc.RequestRef) {
	op.postEventToQueue(msg)
}

func (op *scOperator) IsDismissed() bool {
	op.RLock()
	defer op.RUnlock()
	return op.dismissed
}

func (op *scOperator) dismiss() {
	if op.stopClock != nil {
		op.stopClock()
	}
	op.Lock()
	op.dismissed = true
	close(op.inChan)
	op.Unlock()
}

func (op *scOperator) postEventToQueue(msg interface{}) {
	op.RLock()
	defer op.RUnlock()
	if !op.dismissed {
		op.inChan <- msg
	}
}

func (op *scOperator) dispatchEvent(msg interface{}) {
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

func (op *scOperator) startRoutines() {
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
