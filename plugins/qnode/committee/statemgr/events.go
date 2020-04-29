package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

// triggered whenever new state transaction arrives
func (sm *StateManager) eventStateTransaction(tx sctransaction.Transaction) {

}

// triggered when state update arrives
func (sm *StateManager) eventStateUpdate(stateUpdate state.StateUpdate) {

}
