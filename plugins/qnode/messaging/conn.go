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
	"net"
	"sync"
	"time"
)

type qnodeConnection struct {
	*sync.Mutex
	*buffconn.BufferedConnection
	portAddr      *registry.PortAddr
	runOnce       *sync.Once
	isRunning     bool
	lastHeartbeat time.Time
}

const restartAfter = 10 * time.Second
const dialTimeout = 1 * time.Second

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(500 * time.Millisecond).With(backoff.MaxRetries(1))

func (c *qnodeConnection) handleOutbound() {
	if isInbound(c.portAddr) {
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
	c.Lock()
	c.BufferedConnection = buffconn.NewBufferedConnection(conn)
	c.BufferedConnection.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		c.receiveData(data)
	}))
	c.isRunning = true
	c.Unlock()

	err := c.BufferedConnection.Read()

	if err != nil {
		log.Error(err)
	}

	c.Lock()
	c.BufferedConnection = nil
	c.Unlock()
}

func (c *qnodeConnection) receiveData(data []byte) {
	scid, senderIndex, msgType, msgData, err := unwrapPacket(data)
	if err != nil {
		log.Errorw("msg error", "from", c.portAddr.String(), "err", err)
		return
	}
	oper, ok := getOperator(scid)
	if !ok {
		log.Errorw("message for unexpected sc",
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
	// TODO sending
	return nil
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
