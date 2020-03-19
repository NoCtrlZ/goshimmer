package qserver

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/parameter"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/txdb"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/network/udp"
	"net"
)

type QServer struct {
	udpPort     int
	operators   map[HashValue]*operator.AssemblyOperator
	mockTangle  bool
	mockAddress string
	mockPort    int
	txdb        value.DB
	udpServer   *udp.UDPServer
	Events      serverEvents
}

type serverEvents struct {
	TransactionReceived *events.Event
}

const modulename = "Qserver"

var (
	ServerInstance *QServer
	log            *logger.Logger
)

func StartServer() {
	log = logger.NewLogger(modulename)

	var err error
	err = registry.RefreshAssemblyData()
	if err != nil {
		log.Errorf("StartServer::LoadAllAssemblyData %v", err)
		return
	}

	ServerInstance = &QServer{
		udpPort:     parameter.NodeConfig.GetInt(parameters.UDP_PORT),
		mockTangle:  true,
		mockAddress: parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_IP_ADDR),
		mockPort:    parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PORT),
		txdb:        txdb.NewLocalDb(log),
		udpServer:   createUDPServer(),
		operators:   make(map[HashValue]*operator.AssemblyOperator),
		Events: serverEvents{
			TransactionReceived: events.NewEvent(transactionCaller),
		},
	}
	// ServerInstance events
	ServerInstance.Events.TransactionReceived.Attach(events.NewClosure(transactionEventHandler))

	// setup connection with Value Tangle layer
	value.SetValuetxDB(ServerInstance.txdb)
	value.SetPostFunction(func(vtx value.Transaction) {
		postToValueTangle(ServerInstance, vtx)
	})

	// start UDP server
	addr, port := ServerInstance.GetOwnAddressAndPort()
	err = daemon.BackgroundWorker("Qnode UDP ServerInstance", func(shutdownSignal <-chan struct{}) {
		log.Infof("UDP server listens on %s:%d", addr, port)

		go ServerInstance.udpServer.Listen(addr, port)
		<-shutdownSignal

		log.Infof("UDP server stopped listening...")
	})
	if err != nil {
		log.Errorf("StartServer::daemon.BackgroundWorker %v", err)
		return
	}
	registry.LogLoadedConfigs()
}

func transactionCaller(handler interface{}, params ...interface{}) {
	handler.(func(_ value.Transaction))(params[0].(value.Transaction))
}

// receiving tx from the Value Tangle ontology/layer

func transactionEventHandler(vtx value.Transaction) {
	tx, err := sc.ParseTransaction(vtx)
	if err != nil {
		log.Errorf("%v", err)
		// value tx does not parse to sc.tx. Ignore
		return
	}
	if st, ok := tx.State(); ok {
		// it is state update
		aData, ok := registry.GetAssemblyData(st.AssemblyId())
		if ok {
			// state update has to be processed by this node
			ServerInstance.processState(tx, aData)
		}
	}
	for i, req := range tx.Requests() {
		aid := req.AssemblyId()
		aData, ok := registry.GetAssemblyData(aid)
		if ok {
			// request has to be processed by the node
			reqRef, _ := sc.NewRequestRefFromTx(tx, uint16(i))
			ServerInstance.processRequest(reqRef, aData)
		}
	}
}

func (q *QServer) processState(tx sc.Transaction, assemblyData *registry.AssemblyData) {
	state, _ := tx.State()
	oper, operatorAvailable := ServerInstance.getOperator(state.AssemblyId())
	if operatorAvailable {
		// process config update as normal request
		oper.ReceiveStateUpdate(&sc.StateUpdateMsg{
			Tx: tx,
		})
		return
	}
	oper, err := operator.NewFromState(tx, ServerInstance)
	if err != nil {
		log.Errorf("processState: operator.NewFromState returned: %v", err)
		return
	}
	if oper == nil {
		log.Warnf("processState: this node does not process assembly %s", tx.ShortStr())
		return
	}
	ServerInstance.mustAddOperator(state.AssemblyId(), oper)
	log.Infof("processState: new operator created for aid %s", state.AssemblyId().Short())

}

func (q *QServer) processRequest(reqRef *sc.RequestRef, assemblyData *registry.AssemblyData) {
	req := reqRef.RequestBlock()
	if oper, ok := ServerInstance.getOperator(req.AssemblyId()); ok {
		oper.ReceiveRequest(reqRef)
	}
}

func (q *QServer) isMockTangleAddr(updAddr *net.UDPAddr) bool {
	return q.mockTangle && q.mockPort == updAddr.Port && q.mockAddress == updAddr.IP.String()
}

func (q *QServer) getOperator(aid *HashValue) (*operator.AssemblyOperator, bool) {
	ret, ok := q.operators[*aid]
	if ok && ret.IsDismissed() {
		delete(q.operators, *aid)
		return nil, false
	}
	return ret, ok
}

func (q *QServer) mustAddOperator(aid *HashValue, oper *operator.AssemblyOperator) {
	if _, ok := ServerInstance.operators[*aid]; ok {
		panic(fmt.Sprintf("duplicate operator for aid %s", aid.Short()))
	}
	ServerInstance.operators[*aid] = oper
}

func (q *QServer) IAmInConfig(configData *registry.ConfigData) bool {
	ownIp, ownPort := q.GetOwnAddressAndPort()
	for _, a := range configData.NodeAddresses {
		addr, port := a.AdjustedIP()
		if port == ownPort && addr == ownIp {
			return true
		}
	}
	return false
}

func postToValueTangle(q *QServer, tx value.Transaction) {
	if !q.mockTangle {
		panic("postToValueTangle: not implemented")
	}
	a := net.UDPAddr{
		IP:   net.ParseIP(q.mockAddress),
		Port: q.mockPort,
		Zone: "",
	}
	var buf bytes.Buffer
	err := tx.Encode().Write(&buf)
	if err != nil {
		log.Errorf("%v", err)
		return
	}
	if err = q.SendUDPData(buf.Bytes(), NilHash, MockTangleIdx, 0, &a); err != nil {
		log.Errorf("%v", err)
	}
}
