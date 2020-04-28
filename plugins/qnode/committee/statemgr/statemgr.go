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
	scid        sctransaction.ScId
	isCorrupted bool
	// last state transaction obtained from the tangle
	// it may be not solidified yet in the SC ledger
	lastStateTransaction *sctransaction.Transaction
	// last solid transaction obtained from the tangle
	lastSolidTransaction *sctransaction.Transaction
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

// loads last solid state from the database. Solid state consists of:
// -- last variable state. Loaded from qnode by scid.color
// -- last state update. Loaded from the qnode DB by scid.color and state index, taken from the variable state
// -- last solid transaction. Loaded from the tangle by tx id stored in the last solid state update
// Special conditions for the origin state
func (sm *StateManager) refreshLastSolidState() {
	var err error
	sm.lastSolidVariableState, err = state.LoadVariableState(sm.scid)
	if err != nil {
		log.Errorf("can't load variable state for scid %s: %v", sm.scid.String(), err)
		sm.isCorrupted = true
		return
	}
	solidStateIndex := uint32(0)
	if sm.lastSolidVariableState != nil {
		solidStateIndex = sm.lastSolidVariableState.StateIndex()
	}
	sm.lastSolidStateUpdate, err = state.LoadStateUpdate(sm.scid, solidStateIndex)
	if err != nil {
		log.Errorf("can't load state update index %d for scid %s: %v", solidStateIndex, sm.scid.String(), err)
		sm.isCorrupted = true
		return
	}
	if sm.lastSolidStateUpdate == nil {
		log.Errorf("can't find solid state update with index %d scid %s", solidStateIndex, sm.scid.String())
		return
	}
	if sm.lastSolidVariableState == nil {
		// origin state
		log.Infof("origin state detected for scid %s", sm.scid.String())

		// TODO not finished mess

	}

	//// if state transaction is found, loading last records from the SC ledger db
	//, sm.lastSolidStateUpdate, err = loadLastSolidState(sm.scid)
	//if sm.lastSolidVariableState == nil{
	//	log.Warnf("variable state doesn't exist for scid = %s", sm.scid.String())
	//}
	//if sm.lastSolidVariableState == nil{
	//	log.Warnf("variable state doesn't exist for scid = %s", sm.scid.String())
	//}
	//
	//if sm.lastStateTransaction, err = sctransaction.LoadStateTx(sm.scid); err != nil {
	//	// if there's no state transaction on the tangle it is serious inconsistency
	//	// state manager is supposed to run only when SC is properly initiated
	//
	//	log.Errorf("major inconsistency: Can't get last state transaction for scid %s: %v",
	//		sm.scid.String(), err)
	//	sm.isCorrupted = true
	//
	//	return
	//}
	//
	//if sm.lastSolidVariableState == nil && sm.lastSolidStateUpdate != nil{
	//	if sm.lastSolidStateUpdate.StateIndex() == 0{
	//		log.Infof("origin state detected for scid %s", sm.scid.String())
	//	} else {
	//		log.Errorf("major inconsistency of the state for scid %s", sm.scid.String())
	//
	//		sm.isCorrupted = true
	//		return
	//	}
	//}
	//
	//if err != nil {
	//	log.Errorf("corrupted state: %v", err)
	//	sm.isCorrupted = true
	//	return
	//}
	//if sm.lastSolidVariableState == nil && sm.lastSolidStateUpdate != nil && sm.lastSolidStateUpdate.StateIndex() == 0 {
	//	// origin state
	//	sm.lastSolidVariableState = state.CreateOriginVariableState(sm.lastSolidStateUpdate)
	//	if err := sm.lastSolidVariableState.SaveToDb(); err != nil {
	//		log.Errorf("failed to save origing state for scid = %s", sm.scid.String())
	//	}
	//}
}
