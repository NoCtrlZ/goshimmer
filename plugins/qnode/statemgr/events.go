package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
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

// respond to state update msg.
// It collects state updates while waiting for the anchoring state transaction
// only are stored updates to the current solid variable state
func (sm *StateManager) EventStateUpdateMsg(msg *commtypes.StateUpdateMsg) {
	if !sm.addPendingStateUpdate(msg.StateUpdate) {
		return
	}
	if msg.StateUpdate.StateTransactionId() != sctransaction.NilId && !msg.FromVM {
		// state update has state transaction in it and it is posted by this node as a leader
		// so we need to ask for corresponding state transaction
		sm.asyncLoadStateTransaction(msg.StateUpdate.StateTransactionId(), sm.committee.ScId(), msg.StateUpdate.StateIndex())
		sm.syncMessageDeadline = time.Now().Add(parameters.SyncPeriodBetweenSyncMessages)
	}
	sm.takeAction()
}

// triggered whenever new state transaction arrives
func (sm *StateManager) EventStateTransactionMsg(msg commtypes.StateTransactionMsg) {
	stateBlock, ok := msg.Transaction.State()
	if !ok {
		return
	}
	sm.updateSynchronizationStatus(stateBlock.StateIndex())
	if sm.solidVariableState == nil || stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// only interested for the state transaction to verify latest state update
		return
	}
	sm.nextStateTransaction = msg.Transaction

	sm.takeAction()
}
