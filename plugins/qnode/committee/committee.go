package committee

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commiteeconn"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/statemgr"
)

type Committee struct {
	conn     *commiteeconn.Conn
	stateMgr *statemgr.StateManager
}
