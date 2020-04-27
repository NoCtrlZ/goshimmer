package messaging

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/peering"

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger(modulename)
}
