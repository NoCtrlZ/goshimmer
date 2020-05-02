package consensus

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/consensus"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
