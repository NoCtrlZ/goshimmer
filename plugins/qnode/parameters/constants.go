package parameters

import "time"

const (
	TIMER_ON               = true
	CLOCK_TICK_PERIOD      = 20 * time.Millisecond
	LEADER_ROTATION_PERIOD = 2 * time.Second
)
