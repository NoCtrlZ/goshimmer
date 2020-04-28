// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized with other committee nodes
package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"sync"
)

type StateManager struct {
	sync.RWMutex
	scid sctransaction.ScId
	// last state transaction loaded from the tangle
	lastStateTransaction    *sctransaction.Transaction
	lastVariableState       state.VariableState
	lastEvidencedStateIndex uint32
}

func NewStateManager(scid sctransaction.ScId) *StateManager {
	return &StateManager{
		scid: scid,
	}
}

func (sm *StateManager) isSynchronized() bool {
	if sm.lastStateTransaction == nil {
		return false
	}

	if sm.lastVariableState == nil {
		return false
	}

	if sm.lastEvidencedStateIndex > sm.lastVariableState.StateIndex()+1 {
		return false
	}
}
