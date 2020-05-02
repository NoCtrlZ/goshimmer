package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

// initial loading of the solid state
func (sm *StateManager) initLoadState() {
	var err error

	scid := sm.committee.ScId()
	// load last variable state from the database
	sm.solidVariableState, err = state.LoadVariableState(scid)
	if err != nil {
		log.Errorf("can't load variable state for scid %s: %v", scid.String(), err)

		return
	}
	solidStateIndex := uint32(0)
	if sm.solidVariableState != nil {
		solidStateIndex = sm.solidVariableState.StateIndex()
	}

	// if sm.solidVariableState == nil it may be an origin state

	// load solid state update from db with the state index taken from the variable state
	// state index is 0 if variable state doesn't exist in the DB
	stateUpdate, err := state.LoadStateUpdate(scid, solidStateIndex)
	if err != nil {
		log.Errorf("can't load state update index %d for scid %s: %v", solidStateIndex, scid.String(), err)

		return
	}
	sm.addPendingStateUpdate(stateUpdate)

	// open msg queue for the committee
	sm.committee.SetOperational()

	// here we have at least sm.lastSolidStateUpdate
	// for genesis state sm.solidVariableState == nil
	// async load state transaction
	sm.asyncLoadStateTransaction(stateUpdate.StateTransactionId(), sm.committee.ScId(), stateUpdate.StateIndex())

}

func (sm *StateManager) addPendingStateUpdate(stateUpdate state.StateUpdate) {
	if sm.solidVariableState != nil && stateUpdate.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		return
	}
	var varState state.VariableState
	if sm.solidVariableState == nil {
		varState = state.CreateOriginVariableState(stateUpdate)
	} else {
		varState = sm.solidVariableState.Apply(stateUpdate)
	}
	pending := &pendingStateUpdate{
		stateUpdate:       stateUpdate,
		nextVariableState: varState,
	}
	sm.pendingStateUpdates[hashing.GetHashValue(varState)] = pending
}
