// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized with other committee nodes
package statemgr

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"sync"
)

type StateManager struct {
	sync.RWMutex
	scid        sctransaction.ScId
	isCorrupted bool
	// last state transaction obtained from the tangle
	lastStateTransaction *sctransaction.Transaction
	// last variable state stored in the database
	lastSolidVariableState state.VariableState
	// last state update saved in the database
	lastSolidStateUpdate    state.StateUpdate
	lastEvidencedStateIndex uint32
}

func NewStateManager(scid sctransaction.ScId) *StateManager {
	return &StateManager{
		scid: scid,
	}
}

func (sm *StateManager) isSynchronized() bool {
	if sm.isCorrupted {
		return false
	}
	if sm.lastStateTransaction == nil {
		return false
	}

	if sm.lastSolidVariableState == nil {
		return false
	}

	if sm.lastEvidencedStateIndex > sm.lastSolidVariableState.StateIndex()+1 {
		return false
	}
	return sm.lastStateTransaction.MustState().StateIndex() == sm.lastSolidVariableState.StateIndex()
}

func (sm *StateManager) synchronizationStep() {
	if sm.isCorrupted {
		return
	}
	if sm.isSynchronized() {
		return
	}
	// step 1: get last state transaction
	sm.refreshLastSolidState()
}

func (sm *StateManager) refreshLastSolidState() {
	if sm.lastStateTransaction != nil && sm.lastSolidVariableState != nil {
		return
	}
	if sm.lastStateTransaction == nil || sm.lastSolidVariableState == nil {
		var err error
		if sm.lastStateTransaction, err = sctransaction.LoadStateTx(sm.scid); err != nil {
			log.Errorf("wrong scid or state is corrupted. Can't get last state transaction for scid = %s: %v",
				sm.scid.String(), err)
			sm.isCorrupted = true
			return
		}
		sm.lastSolidVariableState, sm.lastSolidStateUpdate, err = loadLastSolidState(sm.scid)
		if err != nil {
			log.Errorf("corrupted state: %v", err)
			sm.isCorrupted = true
			return
		}
		if sm.lastSolidVariableState == nil && sm.lastSolidStateUpdate != nil && sm.lastSolidStateUpdate.StateIndex() == 0 {
			// origin state
			sm.lastSolidVariableState = state.CreateOriginVariableState(sm.lastSolidStateUpdate)
			if err := sm.lastSolidVariableState.SaveToDb(); err != nil {
				log.Errorf("failed to save origing state for scid = %s", sm.scid.String())
			}
		}
	}
}

func loadLastSolidState(scid sctransaction.ScId) (state.VariableState, state.StateUpdate, error) {
	variableState, err := state.LoadVariableState(scid)
	checkOrigin := false
	if err != nil {
		log.Warnf("no variable state found for scid = %c", scid.String())
		checkOrigin = true // it may be an origin transaction
	}
	var stateUpdate state.StateUpdate
	if checkOrigin {
		if stateUpdate, err = state.LoadStateUpdate(scid, 0); err != nil {
			return nil, nil, fmt.Errorf("failed to load last state for scid = %s", scid.String())
		}
	}
	if variableState == nil && stateUpdate != nil && stateUpdate.StateIndex() != 0 {
		// assertion
		panic("inconsistency")
	}
	// if variableState ==  nil and stateUpdate != nil it is an origin state
	return variableState, stateUpdate, nil
}
