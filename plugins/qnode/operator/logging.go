package operator

import "github.com/iotaledger/hive.go/logger"

const modulename = "operator"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
