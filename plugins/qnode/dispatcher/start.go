package dispatcher

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	qnode_events "github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
)

// start qnode dispatcher daemon worker.
// It serializes all incoming 'ValueTransactionReceived' events
func Start() {
	chTxIn := make(chan *valuetransaction.Transaction)
	processValueTxClosure := events.NewClosure(func(vtx *valuetransaction.Transaction) {
		chTxIn <- vtx
	})

	processPeerMsgClosure := events.NewClosure(func(msg *qnode_events.PeerMessage) {
		if committee := getCommittee(msg.ScColor); committee != nil {
			committee.ReceiveMessage(msg)
		}
	})

	err := daemon.BackgroundWorker("qnode dispatcher", func(shutdownSignal <-chan struct{}) {
		// load all sc data records from registry
		num, err := loadAllSContracts(nil)
		if err != nil || num == 0 {
			log.Error("can't load any SC data from registry. Qnode dispatcher wasn't started")
			return
		}
		log.Debugf("loaded %d SC data record(s) from registry", num)

		// goroutine to serialize incoming value transactions
		go func() {
			for vtx := range chTxIn {
				processIncomingValueTransaction(vtx)
			}
		}()

		<-shutdownSignal

		// starting async cleanup on shutdown
		go func() {
			qnode_events.Events.ValueTransactionReceived.Detach(processValueTxClosure)
			close(chTxIn)
			qnode_events.Events.ValueTransactionReceived.Detach(processPeerMsgClosure)
			log.Infof("qnode dispatcher stopped")
		}()
	})
	if err != nil {
		log.Errorf("failed to initialize qnode dispatcher")
		return
	}
	qnode_events.Events.ValueTransactionReceived.Attach(processValueTxClosure)
	qnode_events.Events.PeerMessageReceived.Attach(processPeerMsgClosure)

	log.Infof("qnode dispatcher started")
}
