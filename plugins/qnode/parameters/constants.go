package parameters

import "time"

const (
	TIMER_ON               = true
	CLOCK_TICK_PERIOD      = 20 * time.Millisecond
	LEADER_ROTATION_PERIOD = 2 * time.Second
	NUM_OLDEST_REQESTS     = 10
	UDP_BUFFER_SIZE        = 2048 * 4
)
