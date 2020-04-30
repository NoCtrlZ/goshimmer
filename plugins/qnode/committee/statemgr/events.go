package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

func (sm *StateManager) EventGetStateUpdateMsg(msg *commtypes.GetStateUpdateMsg) {

}

func (sm *StateManager) EventStateUpdateMsg(msg *commtypes.StateUpdateMsg) {

}

// triggered whenever new state transaction arrives
// it is assumed the transaction has state block
func (sm *StateManager) EventStateTransactionMsg(msg commtypes.StateTransactionMsg) {
	if !sm.isSolidified {
		return
	}
	stateBlock := msg.Transaction.MustState()
	if stateBlock.StateIndex() > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = stateBlock.StateIndex()
	}
	if stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// expected next state index for the solidified state
		// if not, it is out of sync (so, this branch shouldn't happen actually
		return
	}
	// among pending state updates we locate the one, referenced by the new state transaction
	for _, stateUpd := range sm.pendingStateUpdates {
		newVariableState := sm.solidVariableState.Apply(stateUpd)
		if stateBlock.VariableStateHash() == hashing.GetHashValue(newVariableState) {
			// this is the next variable state
			if err := sm.saveStateToDb(newVariableState, stateUpd); err == nil {
				sm.solidVariableState = newVariableState
				sm.lastSolidStateUpdate = stateUpd
				sm.lastSolidStateTransaction = msg.Transaction
				sm.pendingStateUpdates = sm.pendingStateUpdates[:0] // are underlying object GC-ed?
			} else {
				log.Error(err)
			}

			return
		}
	}
	// it comes here when state transaction comes to empty place, where corresponding state updates
	// does not exist. In this case synchronization is needed
	sm.pendingStateUpdates = sm.pendingStateUpdates[:0] // clean up just in case
	sm.isSynchronized = false
}
