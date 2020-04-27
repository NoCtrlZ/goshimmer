package events

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/hive.go/events"
)

var Events struct {
	ValueTransactionReceived *events.Event
	PeerMessageReceived      *events.Event
}

func init() {
	Events.ValueTransactionReceived = events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(func(_ *valuetransaction.Transaction))(params[0].(*valuetransaction.Transaction))
	})
	Events.ValueTransactionReceived = events.NewEvent(func(handler interface{}, params ...interface{}) {
		handler.(func(_ *PeerMessage))(params[0].(*PeerMessage))
	})
}

type PeerMessage struct {
	Timestamp   int64
	ScColor     balance.Color
	SenderIndex uint16
	MsgType     byte
	MsgData     []byte
}
