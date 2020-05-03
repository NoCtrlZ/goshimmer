package vm

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

// TODO

type RuntimeContext struct {
	LeaderPeerIndex uint16
	reqRMsg         *committee.RequestMsg
	ts              time.Time
	stateTx         sctransaction.Transaction
	resultTx        sctransaction.Transaction
	log             *logger.Logger
}
