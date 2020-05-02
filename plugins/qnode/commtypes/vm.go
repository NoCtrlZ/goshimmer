package commtypes

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

// TODO

type RuntimeContext interface {
	ScId() sctransaction.ScId
	Time() time.Time
	PrevTime() time.Time
	Log() *logger.Logger
}
