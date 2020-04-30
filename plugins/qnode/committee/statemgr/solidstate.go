package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
)

// loads and validates last solid state from the database.
// Upon return, if !sm.isCorrupted the last solid state is valid, otherwise SC can't proceed
//
// Solid state consists of:
// -- variable state. Loaded from qnode by scid.color
// -- last state update. Loaded from the qnode DB by scid.color and state index, taken from the variable state
// -- last solid transaction. Loaded from the tangle by tx id stored in the last solid state update
// Special conditions for the origin state. If the function detects non solidified but correct origin state,
// it will solidify it by creating origin variable state in the ledger
func (sm *StateManager) refreshSolidState() {
	if sm.isCorrupted {
		return
	}
	sm.isSolidified = false
	var err error

	// load last variable state from the database
	sm.solidVariableState, err = state.LoadVariableState(sm.scid)
	if err != nil {
		log.Errorf("can't load variable state for scid %s: %v", sm.scid.String(), err)
		sm.isCorrupted = true
		return
	}
	solidStateIndex := uint32(0)
	if sm.solidVariableState != nil {
		solidStateIndex = sm.solidVariableState.StateIndex()
	}

	// if sm.solidVariableState == nil it may be an origin state

	// load solid state update from db with the state index taken from the variable state
	// state index is 0 if variable state doesn't exist in the DB
	sm.lastSolidStateUpdate, err = state.LoadStateUpdate(sm.scid, solidStateIndex)
	if err != nil {
		log.Errorf("can't load state update index %d for scid %s: %v", solidStateIndex, sm.scid.String(), err)
		sm.isCorrupted = true
		return
	}
	if sm.lastSolidStateUpdate == nil {
		log.Errorf("can't find solid state update with index %d scid %s", solidStateIndex, sm.scid.String())
		// not corrupted, but not solid yet
		return
	}
	// load state transaction corresponding to the state update
	sm.lastSolidStateTransaction, err = sctransaction.LoadTx(sm.lastSolidStateUpdate.StateTransactionId())
	if err != nil {
		log.Errorw("major problem: can't load state tx",
			"state index", sm.lastSolidStateUpdate.StateIndex(),
			"tx id", sm.lastSolidStateUpdate.StateTransactionId().String(),
			"scid", sm.scid.String(),
		)

		sm.isCorrupted = true
		return
	}
	// validate state transaction by checking if it has correct state block
	stateBlock, ok := sm.lastSolidStateTransaction.State()
	if !ok || stateBlock.StateIndex() != sm.lastSolidStateUpdate.StateIndex() {
		log.Errorw("major inconsistency: invalid state block in the state transaction",
			"state index", sm.lastSolidStateUpdate.StateIndex(),
			"tx id", sm.lastSolidStateUpdate.StateTransactionId().String(),
			"scid", sm.scid.String(),
		)

		sm.isCorrupted = true
		return
	}
	if sm.solidVariableState != nil {
		// validate the solid variable state and finish the refresh
		if hashing.GetHashValue(sm.solidVariableState) != stateBlock.VariableStateHash() {
			log.Errorw("major problem: last solid state transaction doesn't validate the last solid variable state",
				"state index", sm.lastSolidStateUpdate.StateIndex(),
				"state tx id", sm.lastSolidStateTransaction.String(),
				"scid", sm.scid.String(),
			)

			sm.isCorrupted = true
			return
		}

		sm.isSolidified = true
		return
	}

	if !(sm.solidVariableState == nil && sm.lastSolidStateUpdate != nil) {
		panic("assertion failed: sm.solidVariableState == nil && sm.lastSolidStateUpdate != nil")
	}

	// here sm.solidVariableState == nil, sm.lastSolidStateUpdate != nil
	// so it may be an origin state
	if sm.lastSolidStateUpdate.StateIndex() != 0 {
		log.Errorw("major inconsistency: can't find state block in the state transaction",
			"state index", sm.lastSolidStateUpdate.StateIndex(),
			"tx id", sm.lastSolidStateUpdate.StateTransactionId().String(),
			"scid", sm.scid.String(),
		)

		sm.isCorrupted = true
		return
	}
	// here we have consistent origin state
	// we calculate origin variable state and store it
	sm.solidVariableState = state.CreateOriginVariableState(sm.lastSolidStateUpdate)

	// we have to check if the hash of the origin variable state is equal to the one in the origin transaction

	if hashing.GetHashValue(sm.solidVariableState) != stateBlock.VariableStateHash() {
		// something wrong
		log.Errorw("major inconsistency: origin state transaction is inconsistent with the origin state update",
			"tx id", sm.lastSolidStateTransaction.Id().String(),
			"scid", sm.scid.String(),
		)

		sm.isCorrupted = true
		return
	}
	// save origin state
	if err = sm.solidVariableState.SaveToDb(); err != nil {
		log.Errorw("can't save origin variable state",
			"scid", sm.scid.String(),
			"err", err.Error(),
		)

		sm.isCorrupted = true
		return
	}
	sm.isSolidified = true
}
