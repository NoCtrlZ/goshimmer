package operator

import "github.com/iotaledger/hive.go/logger"

const modulename = "Operator"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
	log.With()
}
