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
	"github.com/iotaledger/hive.go/network"
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
		go connectInboundLoop()

		<-shutdownSignal

		closeAll()

		log.Debugf("stopped qnode communications...")
	}); err != nil {
		panic(err)
	}
}

func OwnPortAddr() *registry.PortAddr {
	return &registry.PortAddr{
		Port: parameter.NodeConfig.GetInt(parameters.QNODE_PORT),
		Addr: "127.0.0.1",
	}
}

func closeAll() {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	for _, cconn := range connections {
		cconn.close()
	}
}

func isInbound(pa *registry.PortAddr) bool {
	return isInboundAddr(pa.String())
}

func isInboundAddr(addr string) bool {
	own := OwnPortAddr().String()
	if own == addr {
		// shouldn't come to this point due to checks before
		panic("can't be same PortAddr")
	}
	return addr < own
}

func addPeerConnection_(portAddr *registry.PortAddr) *qnodeConnection {
	addr := portAddr.String()
	if qconn, ok := connections[addr]; ok {
		return qconn
	}
	connections[addr] = &qnodeConnection{
		RWMutex:       &sync.RWMutex{},
		portAddr:      portAddr,
		lastHeartbeat: time.Now(),
	}
	return connections[addr]
}

func (c *qnodeConnection) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		c.Lock()
		c.runOnce = &sync.Once{}
		c.Unlock()
	}()
}

func connectOutboundLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		connectionsMutex.Lock()
		for _, c := range connections {
			if !isInbound(c.portAddr) {
				c.runOnce.Do(func() {
					go c.runOutbound()
				})
			}
		}
		connectionsMutex.Unlock()
	}
}

func connectInboundLoop() {
	port := parameter.NodeConfig.GetInt(parameters.QNODE_PORT)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Errorf("tcp listen on port %d failed: %v. Restarting connectInboundLoop after 1 sec", port, err)
		go func() {
			time.Sleep(1 * time.Second)
			connectInboundLoop()
		}()
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed accepting a connection request: %v", err)
			continue
		}
		wrongIncoming := false
		addr := conn.RemoteAddr().String()
		if isInboundAddr(addr) {
			wrongIncoming = true
		}
		if !wrongIncoming {
			connectionsMutex.RLock()
			_, ok := connections[addr]
			connectionsMutex.RUnlock()
			if !ok {
				wrongIncoming = true
			}
		}
		if wrongIncoming {
			// connection from (yet) unknown or wrong peer. Drop
			err = conn.Close()
			if err != nil {
				log.Errorf("error while closing during dropping the connection: %v", err)
			} else {
				log.Debugf("dropped incoming connection from unexpected source %s", addr)
			}
			continue
		}
		cconn := connections[addr]
		cconn.Lock()
		if cconn.bufconn != nil {
			log.Panicf("unexpected not nil connection")
		}
		manconn := network.NewManagedConnection(conn)
		cconn.bufconn = buffconn.NewBufferedConnection(manconn)
		cconn.Unlock()
		go cconn.read()
	}
}
