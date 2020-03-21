package mockclientlib

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/parameter"
	"github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/txdb"
	"github.com/iotaledger/hive.go/logger"
)

var db value.DB

func InitMockedValueTangle(log *logger.Logger) {
	locLog := log.Named("mockTangle")
	db = txdb.NewLocalDb(log)
	value.SetValuetxDB(db)
	chPub := make(chan []byte)
	if err := RunPub(parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PUB_TX_PORT), chPub); err != nil {
		locLog.Panic(err)
	}
	value.SetPostFunction(func(vtx value.Transaction) {
		var buf bytes.Buffer
		if err := vtx.Encode().Write(&buf); err != nil {
			locLog.Error(err)
		}
		chPub <- buf.Bytes()
	})
	go listenIncoming(locLog)
}

func listenIncoming(log *logger.Logger) {
	uri := fmt.Sprintf("tcp://%s:%d",
		parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_SERVER_IP_ADDR),
		parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_SERVER_PORT),
	)
	chSub := make(chan []byte)
	go func() {
		for msg := range chSub {
			if vtx, err := value.ParseTransaction(msg); err != nil {
				log.Error(err)
			} else {
				events.Events.TransactionReceived.Trigger(vtx)
			}
		}
	}()
	ReadSub(uri, chSub)
}
