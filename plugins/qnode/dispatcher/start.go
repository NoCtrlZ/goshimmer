package dispatcher

import (
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
)

type qnodeEvents struct {
	ValueTransactionReceived *events.Event
}

var Events qnodeEvents

func init() {
	Events.ValueTransactionReceived = events.NewEvent(ValueTransactionCaller)
}

func ValueTransactionCaller(handler interface{}, params ...interface{}) {
	handler.(func(_ *valuetransaction.Transaction))(params[0].(*valuetransaction.Transaction))
}

// start qnode dispatcher daemon worker.
// It serializes all incoming 'ValueTransactionReceived' events
func Start() {
	chIn := make(chan *valuetransaction.Transaction)
	err := daemon.BackgroundWorker("qnode dispatcher", func(shutdownSignal <-chan struct{}) {
		// serialize incoming value transactions
		go func() {
			for vtx := range chIn {
				processIncomingValueTransaction(vtx)
			}
		}()

		<-shutdownSignal

		// starting acync cleanup on shutdown
		go func() {
			Events.ValueTransactionReceived.DetachAll()
			close(chIn)
			log.Infof("qnode dispatcher stopped")
		}()
	})
	if err != nil {
		log.Errorf("failed to initialize qnode dispatcher")
		return
	}
	Events.ValueTransactionReceived.Attach(events.NewClosure(func(vtx *valuetransaction.Transaction) {
		chIn <- vtx
	}))
	log.Infof("qnode dispatcher started")
}
