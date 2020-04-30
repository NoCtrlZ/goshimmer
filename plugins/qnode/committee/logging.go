package committee

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/committee"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
