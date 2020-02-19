package operator

import (
	"time"
)

// TODO to config parameters
const clockTickPeriod = 20 * time.Millisecond

func (op *AssemblyOperator) getClockTickPeriod() time.Duration {
	op.clockTickPeriodMutex.RLock()
	defer op.clockTickPeriodMutex.RUnlock()
	return op.clockTickPeriod
}

func (op *AssemblyOperator) setClockTickPeriod(period time.Duration) {
	op.clockTickPeriodMutex.Lock()
	defer op.clockTickPeriodMutex.Unlock()
	op.clockTickPeriod = period
}

func (op *AssemblyOperator) startRoutines() {
	// start msg queue routine
	go func() {
		for msg := range op.inChan {
			op.dispatchEvent(msg)
		}
	}()
	// start clock tick routine
	chCancel := make(chan struct{})
	go func() {
		index := 0
		for {
			select {
			case <-chCancel:
				return
			case <-time.After(op.getClockTickPeriod()):
				op.DispatchEvent(timerMsg(index))
				index++
			}
		}
	}()
	op.stopClock = func() {
		close(chCancel)
	}
}
