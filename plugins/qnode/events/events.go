package events

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
)

type qnodeEvents struct {
	TransactionReceived *events.Event
}

const moduleName = "events"

var (
	log    *logger.Logger
	Events qnodeEvents
)

func Init(log1 *logger.Logger) {
	log = log1.Named(moduleName)
	Events.TransactionReceived = events.NewEvent(transactionCaller)
	Events.TransactionReceived.Attach(events.NewClosure(transactionEventHandler))
}

func transactionCaller(handler interface{}, params ...interface{}) {
	handler.(func(_ value.Transaction))(params[0].(value.Transaction))
}

func transactionEventHandler(vtx value.Transaction) {
	tx, err := sc.ParseTransaction(vtx)
	if err != nil {
		log.Errorf("%v", err)
		// value tx does not parse to sc.tx. Ignore
		return
	}
	if st, ok := tx.State(); ok {
		// it is state update
		_, ok := registry.GetAssemblyData(st.SContractId())
		if ok {
			// state update has to be processed by this node
			processState(tx)
		}
	}
	for i, req := range tx.Requests() {
		aid := req.SContractId()
		_, ok := registry.GetAssemblyData(aid)
		if ok {
			// request has to be processed by the node
			reqRef, _ := sc.NewRequestRefFromTx(tx, uint16(i))
			processRequest(reqRef)
		}
	}
}

func processState(tx sc.Transaction) {
	state, _ := tx.State()
	oper, operatorAvailable := messaging.GetOperator(state.SContractId())
	if operatorAvailable {
		// process config update as normal request
		oper.ReceiveStateUpdate(&sc.StateUpdateMsg{
			Tx: tx,
		})
		return
	}
	oper, err := operator.NewFromState(tx)
	if err != nil {
		log.Errorf("processState: operator.NewFromState returned: %v", err)
		return
	}
	if oper == nil {
		log.Warnf("processState: this node does not process assembly %s", tx.ShortStr())
		return
	}
	log.Infof("processState: new operator created for aid %s", state.SContractId().Short())
}

func processRequest(reqRef *sc.RequestRef) {
	req := reqRef.RequestBlock()
	if oper, ok := messaging.GetOperator(req.SContractId()); ok {
		oper.ReceiveRequest(reqRef)
	}
}
