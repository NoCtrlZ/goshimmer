package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/backoff"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"net"
	"sync"
	"time"
)

type qnodeConnection struct {
	sync.Mutex
	*buffconn.BufferedConnection
	portAddr      *registry.PortAddr
	runOnce       *sync.Once
	isRunning     bool
	lastHeartbeat time.Time
}

var (
	connections      map[registry.PortAddr]*qnodeConnection
	connectionsMutex *sync.Mutex
)

func Init() {
	initLogger()
	connections = make(map[registry.PortAddr]*qnodeConnection)
	connectionsMutex = &sync.Mutex{}

	if err := daemon.BackgroundWorker("Qnode connectLoop", func(shutdownSignal <-chan struct{}) {
		log.Debugf("started connectLoop...")

		go connectLoop()
		<-shutdownSignal

		log.Debugf("stopped connectLoop...")
	}); err != nil {
		panic(err)
	}
}

func addConnection(portAddr *registry.PortAddr) bool {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	if _, ok := connections[*portAddr]; ok {
		return false
	}
	connections[*portAddr] = &qnodeConnection{
		Mutex:         sync.Mutex{},
		portAddr:      portAddr,
		lastHeartbeat: time.Now(),
	}
	return true
}

func connectLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		connectionsMutex.Lock()
		for _, c := range connections {
			c.runOnce.Do(func() {
				go c.run()
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

const restartAfter = 10 * time.Second
const dialTimeout = 1 * time.Second

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(500 * time.Millisecond).With(backoff.MaxRetries(1))

func (c *qnodeConnection) run() {
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
	// parse data: assembly id, peer index
}
