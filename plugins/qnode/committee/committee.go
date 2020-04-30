package committee

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/committeeconn"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/consensus"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/statemgr"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

type Committee struct {
	chMsg    chan interface{}
	scdata   *registry.SCData
	conn     *committeeconn.Conn
	stateMgr *statemgr.StateManager
	operator *consensus.Operator
}

func NewCommittee(scdata *registry.SCData) (*Committee, error) {
	conn, err := committeeconn.NewConnection(scdata.ScId, scdata.NodeLocations)
	if err != nil {
		return nil, err
	}
	ret := &Committee{
		chMsg:    make(chan interface{}, 10),
		scdata:   scdata,
		conn:     conn,
		stateMgr: statemgr.NewStateManager(scdata.ScId),
		operator: consensus.NewOperator(),
	}

	go func() {
		for msg := range ret.chMsg {
			ret.dispatchMessage(msg)
		}
	}()

	return ret, nil
}
