package messaging

import (
	"bytes"
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"github.com/iotaledger/hive.go/network"
	"net"
	"sync"
	"time"
)

type qnodeConnection struct {
	*sync.RWMutex
	bufconn             *buffconn.BufferedConnection
	handshakeOutboundOk bool
	portAddr            *registry.PortAddr
	startOnce           *sync.Once
}

const (
	restartAfter = 10 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func (c *qnodeConnection) isInbound() bool {
	return isInboundAddr(c.portAddr.String())
}

func (c *qnodeConnection) runOutbound() {
	if c.isInbound() {
		return
	}
	defer c.runAfter(restartAfter)

	var conn net.Conn
	addr := fmt.Sprintf("%s:%d", c.portAddr.Addr, c.portAddr.Port)
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
	manconn := network.NewManagedConnection(conn)
	bconn := buffconn.NewBufferedConnection(manconn)
	c.setBuferedConnection(bconn)

	if err := c.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	c.readLoopAndClose()
}

func (c *qnodeConnection) sendHandshake() error {
	_, err := c.bufconn.Write([]byte(OwnPortAddr().String()))
	return err
}

func (c *qnodeConnection) setBuferedConnection(bconn *buffconn.BufferedConnection) {
	c.Lock()
	defer c.Unlock()
	c.bufconn = bconn
	c.bufconn.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		c.receiveDataOutbound(data)
	}))
}

func (c *qnodeConnection) close() {
	c.RLock()
	defer c.RUnlock()
	if c.bufconn != nil {
		_ = c.bufconn.Close() // don't check because it may be closed from the other end
		c.bufconn = nil
		c.handshakeOutboundOk = false
	}
}

func (c *qnodeConnection) readLoopAndClose() {
	// read loop. Triggers receive data events for each message
	err := c.bufconn.Read()

	addr := fmt.Sprintf("%s:%d", c.portAddr.Addr, c.portAddr.Port)
	log.Debugf("stopped reading %s", addr)
	if err != nil {
		log.Error(err)
	}
	c.close()
}

func (c *qnodeConnection) receiveDataOutbound(data []byte) {
	if !c.handshakeOutboundOk {
		peerAddr := string(data)
		if peerAddr != c.portAddr.String() {
			log.Error("close the peer connection: wrong handshake message from outbound peer: expected %s got '%s'",
				c.portAddr.String(), peerAddr)
			c.close()
		} else {
			log.Infof("handshake ok with outbound peer %s", peerAddr)
			c.handshakeOutboundOk = true
		}
		return
	}
	c.receiveData(data)
}

func (c *qnodeConnection) receiveData(data []byte) {
	scid, senderIndex, msgType, msgData, err := unwrapPacket(data)
	if err != nil {
		log.Errorw("msg error", "from", c.portAddr.String(), "err", err)
		return
	}
	oper, ok := GetOperator(scid)
	if !ok {
		log.Errorw("message for unexpected scontract",
			"from", c.portAddr.String(),
			"scid", scid.Short(),
			"senderIndex", senderIndex,
			"msgType", msgType,
		)
		return
	}
	if senderIndex >= oper.CommitteeSize() || senderIndex == oper.PeerIndex() {
		log.Errorw("wrong sender index", "from", c.portAddr.String(), "senderIndex", senderIndex)
		return
	}
	if err = oper.ReceiveMsgData(senderIndex, msgType, msgData); err != nil {
		log.Errorw("msg error", "from", c.portAddr.String(), "senderIndex", senderIndex)
	}
}

func (c *qnodeConnection) sendMsgData(data []byte) error {
	c.RLock()
	defer c.RUnlock()

	if c.bufconn == nil {
		return fmt.Errorf("error while sending data: connection with %s not established", c.portAddr.String())
	}
	num, err := c.bufconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes written. err = %v", err)
	}
	return err
}

// returns sc id, sender index, msg type, msg data, error

func unwrapPacket(data []byte) (*HashValue, uint16, byte, []byte, error) {
	rdr := bytes.NewBuffer(data)
	var aid HashValue
	_, err := rdr.Read(aid.Bytes())
	if err != nil {
		return nil, 0, 0, nil, err
	}
	var senderIndex uint16
	err = tools.ReadUint16(rdr, &senderIndex)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	msgType, err := tools.ReadByte(rdr)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return &aid, senderIndex, msgType, rdr.Bytes(), nil
}

func wrapPacket(aid *HashValue, senderIndex uint16, msgType byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.Write(aid.Bytes())
	_ = tools.WriteUint16(&buf, senderIndex)
	buf.WriteByte(msgType)
	buf.Write(data)
	return buf.Bytes()
}
