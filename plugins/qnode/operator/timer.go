package operator

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"time"
)

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
				op.DispatchEvent(timerMsg(index))
				index++
			}
		}
	}()
	op.stopClock = func() {
		close(chCancel)
	}
}
