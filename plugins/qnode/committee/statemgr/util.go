package statemgr

import "github.com/iotaledger/goshimmer/plugins/qnode/state"

// assumes it is consistent, just save into the db
// must an atomic call to badger TODO
func (sm *StateManager) saveStateToDb(varState state.VariableState, stateUpd state.StateUpdate) error {
	if err := varState.SaveToDb(); err != nil {
		return err
	}
	if err := stateUpd.SaveToDb(); err != nil {
		// this is very bad!!!!
		return err
	}
	// ---- end of must be atomic. how to make it?
	log.Infof("variable state #%d has been solidified for scid %s", varState.StateIndex(), sm.scid.String())
	return nil
}
