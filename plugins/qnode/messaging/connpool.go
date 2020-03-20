package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/parameter"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"net"
	"sync"
	"time"
)

type SCOperator interface {
	SContractID() *HashValue
	Quorum() uint16
	CommitteeSize() uint16
	PeerIndex() uint16
	NodeAddresses() []*registry.PortAddr
	ReceiveMsgData(senderIndex uint16, msgType byte, msgData []byte) error
	ReceiveStateUpdate(msg *sc.StateUpdateMsg)
	ReceiveRequest(msg *sc.RequestRef)
	IsDismissed() bool
}

var (
	connections      map[string]*qnodeConnection
	committees       map[HashValue]*CommitteeConn
	connectionsMutex *sync.RWMutex
)

func Init() {
	initLogger()
	connections = make(map[string]*qnodeConnection)
	committees = make(map[HashValue]*CommitteeConn)
	connectionsMutex = &sync.RWMutex{}

	if err := daemon.BackgroundWorker("Qnode connectOutboundLoop", func(shutdownSignal <-chan struct{}) {
		log.Debugf("started connectOutboundLoop...")

		go connectOutboundLoop()
		<-shutdownSignal

		log.Debugf("stopped connectOutboundLoop...")
	}); err != nil {
		panic(err)
	}
}

func ownPortAddr() *registry.PortAddr {
	return &registry.PortAddr{
		Port: parameter.NodeConfig.GetInt(parameters.QNODE_PORT),
		Addr: "127.0.0.1",
	}
}

func isInbound(pa *registry.PortAddr) bool {
	own := ownPortAddr()
	switch {
	case pa.Addr < own.Addr:
		return true
	case pa.Addr > own.Addr:
		return false
	}
	if pa.Port == own.Port {
		panic("can't be same PortAddr")
	}
	return pa.Port < own.Port
}

func addPeerConnection_(portAddr *registry.PortAddr) *qnodeConnection {
	addr := portAddr.String()
	if qconn, ok := connections[addr]; ok {
		return qconn
	}
	connections[addr] = &qnodeConnection{
		Mutex:         &sync.Mutex{},
		portAddr:      portAddr,
		lastHeartbeat: time.Now(),
	}
	return connections[addr]
}

func connectOutboundLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		connectionsMutex.Lock()
		for _, c := range connections {
			c.runOnce.Do(func() {
				go c.handleOutbound()
			})
		}
		connectionsMutex.Unlock()
	}
}

func (c *qnodeConnection) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		c.Lock()
		c.runOnce = &sync.Once{}
		c.Unlock()
	}()
}

func connectInboundLoop() {
	port := parameter.NodeConfig.GetInt(parameters.QNODE_PORT)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicf("tcp listen on port %d failed: %v", port, err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed accepting a connection request: %v", err)
			continue
		}
		addr := conn.RemoteAddr().String()
		connectionsMutex.RLock()
		cconn, ok := connections[addr]
		connectionsMutex.RUnlock()
		if !ok {
			// connection from yet unknown. Drop
			err = conn.Close()
			if err != nil {
				log.Errorf("error while closing connection: %v", err)
			} else {
				log.Debugf("dropped incoming connection from unexpected source %s", addr)
			}
			continue
		}
		cconn.Lock()
		if cconn.BufferedConnection != nil {
			log.Panicf("unexpected not nil connection")
		}
		cconn.BufferedConnection = buffconn.NewBufferedConnection(conn)
		cconn.Unlock()
	}
}
