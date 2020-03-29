package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/backoff"
	"net"
	"sync"
	"time"
)

type qnodePeer struct {
	*sync.RWMutex
	peerconn     *peeredConnection // nil mean not connected
	handshakeOk  bool
	peerPortAddr *registry.PortAddr
	startOnce    *sync.Once
}

const (
	restartAfter = 1 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func (c *qnodePeer) isInbound() bool {
	return isInboundAddr(c.peerPortAddr.String())
}

func (c *qnodePeer) connStatus() (bool, bool) {
	c.RLock()
	defer c.RUnlock()
	return c.peerconn != nil, c.handshakeOk
}

func (c *qnodePeer) closeConn() {
	c.Lock()
	defer c.Unlock()
	if c.peerconn != nil {
		_ = c.peerconn.Close()
	}
}

func (c *qnodePeer) runOutbound() {
	if c.peerconn != nil {
		panic("c.peerconn != nil")
	}
	if c.isInbound() {
		return
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

func (c *qnodePeer) sendHandshake() error {
	num, err := c.peerconn.Write([]byte(OwnPortAddr().String()))
	log.Debugf("sendHandshake %d bytes to %s", num, c.peerPortAddr.String())
	return err
}

func (c *qnodePeer) receiveData(data []byte) {
	scid, senderIndex, msgType, msgData, err := unwrapPacket(data)
	if err != nil {
		log.Errorw("msg error", "from", c.peerPortAddr.String(), "err", err)
		return
	}
	committee, ok := getCommittee(scid)
	if !ok {
		log.Errorw("message for unexpected scontract",
			"from", c.peerPortAddr.String(),
			"scid", scid.Short(),
			"senderIndex", senderIndex,
			"msgType", msgType,
		)
		return
	}
	if senderIndex >= committee.operator.CommitteeSize() || senderIndex == committee.operator.PeerIndex() {
		log.Errorw("wrong sender index", "from", c.peerPortAddr.String(), "senderIndex", senderIndex)
		return
	}
	committee.recvDataCallback(senderIndex, msgType, msgData)
}

func (c *qnodePeer) sendMsgData(data []byte) error {
	c.RLock()
	defer c.RUnlock()

	if c.peerconn == nil {
		return fmt.Errorf("error while sending data: connection with %s not established", c.peerPortAddr.String())
	}
	num, err := c.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes written. err = %v", err)
	}
	return err
}