package mockclientlib

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/txdb"
	"github.com/iotaledger/hive.go/logger"
)

var db value.DB

func InitMockedValueTangle(log *logger.Logger) {
	db = txdb.NewLocalDb(log)
	value.SetValuetxDB(db)
	value.SetPostFunction(func(vtx value.Transaction) {
	})
	go listenIncoming()
}

func listenIncoming() {

}
