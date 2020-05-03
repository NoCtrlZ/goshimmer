package commiteeimpl

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/committeeObj"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
