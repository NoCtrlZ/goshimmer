package messaging

import (
	"fmt"
	qnode_events "github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/backoff"
	"net"
	"sync"
	"time"
)

// represents point-to-point TCP connection between two qnodes and another
// it is used as transport for message exchange
// Another end is always using the same connection
// the Peer takes care about exchanging heartbeat messages.
// It keeps last several received heartbeats as "lad" data to be able to calculate how synced/unsynced
// clocks of peer are.
type Peer struct {
	*sync.RWMutex
	peerconn     *peeredConnection // nil means not connected
	handshakeOk  bool
	peerPortAddr *registry.PortAddr
	startOnce    *sync.Once
	numUsers     int
	// heartbeats and latencies
	lastHeartbeatReceived time.Time
	lastHeartbeatSent     time.Time
	latencyRingBuf        [numHeartbeatsToKeep]int64
	hbRingBufIdx          int
}

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func (c *Peer) isInbound() bool {
	return isInboundAddr(c.peerPortAddr.String())
}

func (c *Peer) connStatus() (bool, bool) {
	c.RLock()
	defer c.RUnlock()
	return c.peerconn != nil, c.handshakeOk
}

func (c *Peer) closeConn() {
	c.Lock()
	defer c.Unlock()
	if c.peerconn != nil {
		_ = c.peerconn.Close()
	}
}

// dials outbound address and established connection
func (c *Peer) runOutbound() {
	if c.isInbound() {
		return
	}
	if c.peerconn != nil {
		panic("c.peerconn != nil")
	}
	log.Debugf("runOutbound %s", c.peerPortAddr.String())

	defer c.runAfter(restartAfter)

	var conn net.Conn
	addr := fmt.Sprintf("%s:%d", c.peerPortAddr.Addr, c.peerPortAddr.Port)
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return fmt.Errorf("dial %s failed: %w", addr, err)
		}
		return nil
	}); err != nil {
		log.Error(err)
		return
	}
	//manconn := network.NewManagedConnection(conn)
	c.peerconn = newPeeredConnection(conn, c)
	if err := c.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	log.Debugf("starting reading outbound %s", c.peerPortAddr.String())
	if err := c.peerconn.Read(); err != nil {
		log.Error(err)
	}
	log.Debugf("stopped reading. Closing %s", c.peerPortAddr.String())
	c.closeConn()
}

// sends handshake message. It contains IP address of this end.
// The address is used by another end for peering
func (c *Peer) sendHandshake() error {
	data, _ := encodeMessage(&qnode_events.PeerMessage{
		MsgType: MsgTypeHandshake,
		MsgData: []byte(OwnPortAddr().String()),
	})
	num, err := c.peerconn.Write(data)
	log.Debugf("sendHandshake %d bytes to %s", num, c.peerPortAddr.String())
	return err
}

func (c *Peer) sendData(data []byte) error {
	c.RLock()
	defer c.RUnlock()

	if c.peerconn == nil {
		return fmt.Errorf("error while sending data: connection with %s not established", c.peerPortAddr.String())
	}
	num, err := c.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes written. err = %v", err)
	}
	go c.scheduleNexHeartbeat()
	return err
}
