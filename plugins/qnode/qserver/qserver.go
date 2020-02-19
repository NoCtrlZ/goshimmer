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
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/network/udp"
	"net"
)

type QServer struct {
	udpPort      int
	OperatorData map[HashValue]*operator.AssemblyData
	operators    map[HashValue]*operator.AssemblyOperator
	mockTangle   bool
	mockAddress  string
	mockPort     int
	udpServer    *udp.UDPServer
	Events       serverEvents
}

type serverEvents struct {
	NodeEvent *events.Event
}

const modulename = "qserver"

var (
	ServerInstance *QServer
	log            *logger.Logger
)

func StartServer() error {
	log = logger.NewLogger(modulename)

	var opdata map[HashValue]*operator.AssemblyData
	var err error
	opdata, err = operator.LoadAllOperatorData()
	if err != nil {
		return err
	}
	ServerInstance = &QServer{
		udpPort:      parameter.NodeConfig.GetInt(parameters.UDP_PORT),
		mockTangle:   true,
		mockAddress:  parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_IP_ADDR),
		mockPort:     parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PORT),
		OperatorData: opdata,
		udpServer:    createUDPServer(),
		operators:    make(map[HashValue]*operator.AssemblyOperator),
		Events: serverEvents{
			NodeEvent: events.NewEvent(nodeEventCaller),
		},
	}
	// ServerInstance events
	ServerInstance.Events.NodeEvent.Attach(events.NewClosure(nodeEventHandler))
	addr, port := ServerInstance.GetOwnAddressAndPort()
	err = daemon.BackgroundWorker("Qnode UDP ServerInstance", func(shutdownSignal <-chan struct{}) {
		log.Infof("UDP server listens on %s:%d", addr, port)

		go ServerInstance.udpServer.Listen(addr, port)
		<-shutdownSignal

		log.Infof("UDP server stopped listening...")
	})
	if err != nil {
		return err
	}
	logLoadedConfigs()
	return nil
}

func logLoadedConfigs() {
	log.Debugf("loaded %d operator data record(s)", len(ServerInstance.OperatorData))
	for _, od := range ServerInstance.OperatorData {
		log.Debugf("aid = %s dscr = '%s'", od.AssemblyId.String(), od.Description)
	}
}

func nodeEventCaller(handler interface{}, params ...interface{}) {
	handler.(func(_ value.Transaction))(params[0].(value.Transaction))
}

// receiving tx from the Value Tangle ontology/layer

func nodeEventHandler(txval value.Transaction) {
	log.Info("nodeEventHandler")
	tx, err := sc.ParseTransaction(txval)
	if err != nil {
		// value tx does not parse to sc tx. Ignore
		return
	}
	if st, ok := tx.State(); ok {
		// it is state update
		aData := ServerInstance.findAssemblyData(st.AssemblyId())
		if aData != nil {
			// state update has to be processed by this node
			ServerInstance.processState(tx, aData)
		}
	}
	for i, req := range tx.Requests() {
		aid := req.AssemblyId()
		aData := ServerInstance.findAssemblyData(aid)
		if aData != nil {
			// request has to be processed by the node
			ServerInstance.processRequest(tx, uint16(i), aData)
		}
	}
}

func (q *QServer) processState(tx sc.Transaction, assemblyData *operator.AssemblyData) {
	state, _ := tx.State()
	oper, operatorAvailable := ServerInstance.getOperator(state.AssemblyId())
	if operatorAvailable {
		// process config update as normal request
		oper.DispatchEvent(&sc.StateUpdateMsg{
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
		log.Errorf("processState: this node does not process assembly %s", tx.ShortStr())
		return
	}
	ServerInstance.mustAddOperator(state.AssemblyId(), oper)
	log.Errorf("processState: new operator created for aid %s", state.AssemblyId().Short())

}

func (q *QServer) processRequest(tx sc.Transaction, reqIndex uint16, assemblyData *operator.AssemblyData) {
	req := tx.Requests()[reqIndex]
	if oper, ok := ServerInstance.getOperator(req.AssemblyId()); ok {
		oper.DispatchEvent(&sc.RequestMsg{
			Tx:           tx,
			RequestIndex: reqIndex,
		})
	}
}

func (q *QServer) isMockTangleAddr(updAddr *net.UDPAddr) bool {
	return q.mockTangle && q.mockPort == updAddr.Port && q.mockAddress == updAddr.IP.String()
}

func (q *QServer) findAssemblyData(aid *HashValue) *operator.AssemblyData {
	if ret, ok := q.OperatorData[*aid]; ok {
		return ret
	}
	return nil
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

func (q *QServer) IAmInConfig(configData *operator.ConfigData) bool {
	ownIp, ownPort := q.GetOwnAddressAndPort()
	for _, a := range configData.OperatorAddresses {
		addr, port := a.AdjustedIP()
		if port == ownPort && addr == ownIp {
			return true
		}
	}
	return false
}

// Comm interface

func (q *QServer) GetOwnAddressAndPort() (string, int) {
	return "127.0.0.1", q.udpPort
}

func (q *QServer) SendUDPData(data []byte, aid *HashValue, senderIndex uint16, msgType byte, addr *net.UDPAddr) error {
	wrapped := WrapUDPPacket(aid, senderIndex, msgType, data)
	_, err := q.udpServer.GetSocket().WriteTo(wrapped, addr)
	return err
}

func (q *QServer) PostToValueTangle(tx value.Transaction) error {
	if q.mockTangle {
		a := net.UDPAddr{
			IP:   net.ParseIP(q.mockAddress),
			Port: q.mockPort,
			Zone: "",
		}
		data := mustEncodeTx(tx)
		return q.SendUDPData(data, NilHash, MockTangleIdx, 0, &a)
	}
	panic("PostToValueTangle: not implemented")
}

func mustEncodeTx(tx value.Transaction) []byte {
	var buf bytes.Buffer
	err := tx.Encode().Write(&buf)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}