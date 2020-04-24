package dispatcher

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/dispatcher"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
