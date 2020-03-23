package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/parameter"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
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
		RWMutex:      &sync.RWMutex{},
		peerPortAddr: portAddr,
		startOnce:    &sync.Once{},
	}
	log.Debugf("added new connection %s inbound = %v", addr, connections[addr].isInbound())
	return connections[addr]
}

func (c *qnodeConnection) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		c.Lock()
		c.startOnce = &sync.Once{}
		c.Unlock()
		log.Debugf("will run %s again", c.peerPortAddr.String())
	}()
}

func connectOutboundLoop() {
	count := 0
	numConnected := 0
	for {
		time.Sleep(100 * time.Millisecond)
		numConnected = 0
		count++
		sumUp := (count % 50) == 0
		connectionsMutex.Lock()
		for _, c := range connections {
			c.startOnce.Do(func() {
				go c.runOutbound()
			})
			if sumUp && c.isConnected() {
				numConnected++
			}
		}
		connectionsMutex.Unlock()
		if sumUp {
			log.Debugf("number of connected peers: %d", numConnected)
		}
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
		log.Debugf("accepted connection from %s", conn.RemoteAddr().String())

		manconn := network.NewManagedConnection(conn)
		bconn := buffconn.NewBufferedConnection(manconn)
		bconn.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
			receiveHandshakeInbound(bconn, data)
		}))
		bconn.Events.Close.Attach(events.NewClosure(func() {
			log.Errorf("inbound closed during handshake %s", conn.RemoteAddr().String())
		}))
		go func() {
			log.Debugf("starting reading inbound %s", conn.RemoteAddr().String())
			if err := bconn.Read(); err != nil {
				log.Error(err)
			}
			_ = bconn.Close()
		}()
	}
}

func receiveHandshakeInbound(bconn *buffconn.BufferedConnection, data []byte) {
	peerAddr := string(data)
	log.Debugf("received handshake inbound %s", peerAddr)

	connectionsMutex.RLock()
	cconn, ok := connections[peerAddr]
	connectionsMutex.RUnlock()

	if !ok || !cconn.isInbound() {
		log.Errorf("inbound connection from unexpected peer %s. Closing..", peerAddr)
		_ = bconn.Close()
		return
	}
	cconn.setBufferedConnection(bconn, func(data []byte) {
		cconn.receiveData(data)
	})
	if err := cconn.sendHandshake(); err != nil {
		log.Errorf("error while sending handshake: %v", err)
		cconn.close()
		return
	}
	log.Infof("connected inbound %s", peerAddr)
}
