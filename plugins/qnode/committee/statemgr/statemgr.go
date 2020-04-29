// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized with other committee nodes
package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

type StateManager struct {
	scid sctransaction.ScId

	// state is corrupted, SC can't proceed
	isCorrupted bool

	// pending state updates are state updates calculated by the VM
	// may be several of them with different timestamps
	// upon arrival of state transaction one of them will be solidified
	pendingStateUpdates []state.StateUpdate

	// last solid transaction obtained from the tangle by the reference from the
	// solid state update
	lastSolidStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	solidVariableState state.VariableState

	// last state update stored in the database. Obtained by the index stored in variable state
	// In case of origin state index is 0
	lastSolidStateUpdate state.StateUpdate

	// last state transaction received from the tangle
	// it may be not solidified yet in the SC ledger
	// if it coincides with the lastSolidStateTransaction, the state is in sync, otherwise it is not
	lastStateTransaction *sctransaction.Transaction

	// largest state index seen from other messages. If this index is more than 1 step ahead then
	// the solid one, state is not synced
	lastEvidencedStateIndex uint32
}

func NewStateManager(scid sctransaction.ScId) *StateManager {
	return &StateManager{
		scid:                scid,
		pendingStateUpdates: make([]state.StateUpdate, 0),
	}
}

func (sm *StateManager) isSynchronized() bool {
	if sm.isCorrupted {
		return false
	}
	if sm.lastStateTransaction == nil {
		return false
	}

	if sm.solidVariableState == nil {
		return false
	}

	if sm.lastEvidencedStateIndex > sm.solidVariableState.StateIndex()+1 {
		return false
	}
	if sm.lastStateTransaction.MustState().StateIndex() != sm.solidVariableState.StateIndex() {
		return false
	}
	return true
}

func (sm *StateManager) synchronizationStep() {
	if sm.isCorrupted {
		return
	}
	if sm.isSynchronized() {
		return
	}
	// step 1: get last state transaction
	sm.refreshSolidState()
}
