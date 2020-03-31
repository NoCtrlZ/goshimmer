package messaging

import (
	"encoding/binary"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

// message type is 1 byte
// from 0 until maxSpecMsgCode inclusive it is reserved for heartbeat and other message types
// these messages are processed by receiveSpecMsg method
// the rest are forwarded to SC operator

const (
	FIRST_SC_MSG_TYPE = byte(16)

	// heartbeat msg
	HEARTBEAT_MSG_TYPE = byte(1)
	lastHB             = 5               // number of heartbeats to save for average latency
	heartbeatEvery     = 5 * time.Second // heartBeat period
	isDeadAfterMissing = 2               // is dead after 4 heartbeat periods missing
)

func (c *qnodePeer) receiveConnMsg(msgType byte, data []byte) {
	switch msgType {
	case HEARTBEAT_MSG_TYPE:
		c.receiveHeartbeat(data)
	default:
		log.Errorf("wrong spec. communication message from %s", c.peerPortAddr.String())
	}
}

func (c *qnodePeer) receiveHeartbeat(data []byte) {
	if len(data) != 8 {
		log.Error("expected 8 bytes for heartbeat")
		return
	}
	ts := int64(binary.LittleEndian.Uint64(data))

	c.hbMutex.Lock()
	c.lastHeartbeat = time.Now()
	latency := c.lastHeartbeat.UnixNano() - ts
	c.latency[c.latencyIdx] = latency
	c.latencyIdx = (c.latencyIdx + 1) % lastHB
	c.hbMutex.Unlock()

	//log.Debugf("heartbeat received from %s, latency %d nanosec", c.peerPortAddr.String(), latency)
}

func (c *qnodePeer) sendHeartbeat() {
	ts := tools.Uint64To8Bytes(uint64(time.Now().UnixNano()))
	wrapped := wrapPacket(nil, 0, HEARTBEAT_MSG_TYPE, ts)
	if err := c.sendMsgData(wrapped); err == nil {
		//log.Debugf("heartbeat sent to %s", c.peerPortAddr.String())
	}
	// ignore errors
}

// return true if is alive and average latency in nanosec
func (c *qnodePeer) isAlive() (bool, int64) {
	c.RLock()
	defer c.RUnlock()
	if c.peerconn == nil {
		return false, 0
	}
	c.hbMutex.RLock()
	defer c.hbMutex.RUnlock()

	if time.Since(c.lastHeartbeat) > heartbeatEvery*isDeadAfterMissing {
		return false, 0
	}
	sum := int64(0)
	for _, l := range c.latency {
		sum += l
	}
	return true, sum / lastHB
}

func (c *qnodePeer) startHeartbeat() func() {
	log.Debugf("start heartbeat sending to %s", c.peerPortAddr.String())

	chCancel := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(heartbeatEvery):
				c.sendHeartbeat()
			case <-chCancel:
				log.Debugf("stopped heartbeat sending to %s", c.peerPortAddr.String())
				return
			}
		}
	}()
	return func() {
		close(chCancel)
	}
}
