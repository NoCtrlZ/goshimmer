package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"time"
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
	sm.addPendingStateUpdate(msg.StateUpdate)
	sm.takeAction()
}

// triggered whenever new state transaction arrives
// it is assumed the transaction has state block
func (sm *StateManager) EventStateTransactionMsg(msg commtypes.StateTransactionMsg) {
	stateBlock := msg.Transaction.MustState()
	sm.accountLargestStateIndex(stateBlock.StateIndex())

	if stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// only interested for the state transaction to verify latest state update
		return
	}

	sm.nextStateTransaction = msg.Transaction
	sm.stateTransactionArrived = time.Now()

	sm.takeAction()
}
