package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

// respond to sync request 'GetStateUpdate'
func (sm *StateManager) EventGetStateUpdateMsg(msg *commtypes.GetStateUpdateMsg) {
	if stateUpd, err := state.LoadStateUpdate(sm.committee.ScId(), msg.StateIndex); err == nil {
		_ = sm.committee.SendMsg(msg.SenderIndex, commtypes.MsgStateUpdate, hashing.MustBytes(&commtypes.StateUpdateMsg{
			StateUpdate: stateUpd,
		}))
	}
	// no need for action because it doesn't change of the state
}

// react to state update msg.
// It collects state updates while waiting for the anchoring state transaction
// only are stored updates to the current solid variable state
func (sm *StateManager) EventStateUpdateMsg(msg *commtypes.StateUpdateMsg) {
	if msg.StateUpdate.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// only interested in the state updates for the current solid variable state
		return
	}
	sm.pendingStateUpdates = append(sm.pendingStateUpdates, msg.StateUpdate)
	sm.takeAction()
}

// triggered whenever new state transaction arrives
// it is assumed the transaction has state block
func (sm *StateManager) EventStateTransactionMsg(msg commtypes.StateTransactionMsg) {
	if !sm.isSolidified {
		return
	}
	stateBlock := msg.Transaction.MustState()
	sm.accountLargestStateIndex(stateBlock.StateIndex())

	if stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// only interested for the state transaction to verify latest state update
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
	sm.setSynchronized(false)
}
