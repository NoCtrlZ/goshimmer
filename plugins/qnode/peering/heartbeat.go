package messaging

import (
	"time"
)

// message type is 1 byte
// from 0 until maxSpecMsgCode inclusive it is reserved for heartbeat and other message types
// these messages are processed by processHeartbeat method
// the rest are forwarded to SC operator

func (c *Peer) initHeartbeats() {
	c.lastHeartbeatSent = time.Time{}
	c.lastHeartbeatReceived = time.Time{}
	c.hbRingBufIdx = 0
	for i := range c.latencyRingBuf {
		c.latencyRingBuf[i] = 0
	}
}

func (c *Peer) receiveHeartbeat(ts int64) {
	c.Lock()
	c.lastHeartbeatReceived = time.Now()
	lagNano := c.lastHeartbeatReceived.UnixNano() - ts
	c.latencyRingBuf[c.hbRingBufIdx] = lagNano
	c.hbRingBufIdx = (c.hbRingBufIdx + 1) % numHeartbeatsToKeep
	c.Unlock()

	//log.Debugf("heartbeat received from %s, lag %f milisec", c.peerPortAddr.String(), float64(lagNano/10000)/100)
}

func (c *Peer) scheduleNexHeartbeat() {
	time.Sleep(heartbeatEvery)
	if peerAlive, _ := c.isAlive(); !peerAlive {
		log.Debugf("stopped sending heartbeat: peer %s is dead", c.peerPortAddr.String())
		return
	}

	c.Lock()

	if time.Since(c.lastHeartbeatSent) < heartbeatEvery {
		// was recently sent. exit
		c.Unlock()
		return
	}
	var hbMsgData []byte
	hbMsgData, c.lastHeartbeatSent = wrapPacket(nil)

	c.Unlock()

	_ = c.sendData(hbMsgData)
	//log.Debugf("sent heartbeat to %s", c.peerPortAddr.String())

	// repeat after some time
}

// return true if is alive and average latencyRingBuf in nanosec
func (c *Peer) isAlive() (bool, int64) {
	c.RLock()
	defer c.RUnlock()
	if c.peerconn == nil || !c.handshakeOk {
		return false, 0
	}

	if time.Since(c.lastHeartbeatReceived) > heartbeatEvery*isDeadAfterMissing {
		return false, 0
	}
	sum := int64(0)
	for _, l := range c.latencyRingBuf {
		sum += l
	}
	return true, sum / numHeartbeatsToKeep
}
