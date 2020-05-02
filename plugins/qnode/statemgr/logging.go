package statemgr

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/statemgr"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
