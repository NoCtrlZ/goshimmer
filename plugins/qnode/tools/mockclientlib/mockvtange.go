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

var (
	db  value.DB
	log *logger.Logger
)

func InitLogger() {
	log = logger.NewLogger("mockTangle")
}

func InitMockedValueTangle() {
	db = txdb.NewLocalDb(log)
	value.SetValuetxDB(db)
	chPub := make(chan []byte)
	pubPort := parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PUB_TX_PORT)
	if err := RunPub(pubPort, chPub); err != nil {
		log.Panic(err)
	}
	log.Infof("will be publishing txs to mocked tangle over port %d", pubPort)

	value.SetPostFunction(func(vtx value.Transaction) {
		var buf bytes.Buffer
		if err := vtx.Encode().Write(&buf); err != nil {
			log.Error(err)
		}
		chPub <- buf.Bytes()
	})
	go listenIncoming()
}

func listenIncoming() {
	uri := fmt.Sprintf("tcp://%s:%d",
		parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_SERVER_IP_ADDR),
		parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_SERVER_PORT),
	)
	chSub := make(chan InMessage)
	go func() {
		for msg := range chSub {
			if vtx, err := value.ParseTransaction(msg.Data); err != nil {
				log.Error(err)
			} else {
				err = db.PutTransaction(vtx)
				if err != nil {
					log.Error(err)
				}
				events.Events.TransactionReceived.Trigger(vtx)
			}
		}
	}()
	ReadSub(uri, chSub)
}
