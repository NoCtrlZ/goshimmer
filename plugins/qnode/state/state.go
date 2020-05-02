package state

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/labstack/gommon/log"
)

// assumes it is consistent, just save into the db
// must an atomic call to badger TODO
func SaveStateToDb(stateUpd StateUpdate, varState VariableState, reqIds *[]sctransaction.RequestId) error {
	if err := varState.SaveToDb(); err != nil {
		return err
	}
	if err := stateUpd.SaveToDb(); err != nil {
		// this is very bad!!!!
		return err
	}

	if err := MarkReqIdsProcessed(reqIds); err != nil {
		return err
	}
	// ---- end of must be atomic. how to make it?
	scid := stateUpd.ScId()
	log.Infof("variable state #%d has been solidified for scid %s", varState.StateIndex(), scid.String())
	return nil
}

func MarkReqIdsProcessed(reqIds *[]sctransaction.RequestId) error {
	panic("implement me")
}
