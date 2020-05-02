package statemgr

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"time"
)

func (sm *StateManager) takeAction() {
	if sm.checkStateTransition() {
		return
	}
	sm.queryStateUpdateFromPeerIfNeeded()
}

// if state is corrupted, will never synchronize
func (sm *StateManager) checkStateTransition() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state updates we locate the one, consistent with the next state transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	pending, ok := sm.pendingStateUpdates[varStateHash]
	if !ok {
		// corresponding state update wasn't found among pending state updates
		return false
	}
	// found corresponding pending state
	// it is approved by the nextStateTransaction
	pending.stateUpdate.SetStateTransactionId(sm.nextStateTransaction.Id())

	if err := sm.saveStateToDb(pending.nextVariableState, pending.stateUpdate); err != nil {
		log.Errorf("failed to save state #%d: %v", pending.stateUpdate.StateIndex(), err)
		return false
	}

	prevStateIndex := ""
	if sm.solidVariableState.StateIndex() > 0 {
		prevStateIndex = fmt.Sprintf("#%d", sm.solidVariableState.StateIndex()-1)
	}
	log.Infof("state transition %s --> #%d scid %s", prevStateIndex, sm.solidVariableState.StateIndex())

	sm.solidVariableState = pending.nextVariableState
	sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), varStateHash.Bytes())
	sm.permutationIndex = 0
	sm.syncMessageDeadline = time.Now() // if not synced the immediately
	return false
}

func (sm *StateManager) queryStateUpdateFromPeerIfNeeded() {
	if sm.solidVariableState == nil {
		return
	}
	if sm.isSynchronized() {
		// state is synced
		return
	}
	// not synced
	if !sm.syncMessageDeadline.Before(time.Now()) {
		// not time yet
		return
	}
	// it is time to ask for the next state update to next peer in the permutation
	sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
	data := hashing.MustBytes(&commtypes.GetStateUpdateMsg{
		StateIndex: sm.solidVariableState.StateIndex() + 1,
	})
	for i := uint16(0); i < sm.committee.Size(); i++ {
		targetPeerIndex := sm.permutationOfPeers[sm.permutationIndex]
		if err := sm.committee.SendMsg(targetPeerIndex, commtypes.MsgGetStateUpdate, data); err == nil {
			break
		}
		sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
		sm.syncMessageDeadline = time.Now().Add(parameters.SyncPeriodBetweenSyncMessages)
	}
}

func (sm *StateManager) checkSynchronized(idx uint32) {
	// synced state is when current state index is behind
	// the largestEvidencedStateIndex no more than by 1 point
	wasSynchronized := sm.isSynchronized()
	if idx > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = idx
	}
	if !sm.isSynchronized() && wasSynchronized {
		sm.syncMessageDeadline = time.Now()
	}
}

func (sm *StateManager) isSynchronized() bool {
	return sm.largestEvidencedStateIndex-sm.solidVariableState.StateIndex() <= 1
}

// async loads state transaction from DB and validates it
// posts 'StateTransactionMsg' to the committee upon success
func (sm *StateManager) asyncLoadStateTransaction(txid transaction.Id, scid sctransaction.ScId, stateIndex uint32) {
	go func() {
		tx, err := sctransaction.LoadTx(txid)
		if err != nil {
			log.Errorf("can't load state tx",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		stateBlock, ok := tx.State()
		if !ok {
			log.Errorf("not a state tx",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		if *stateBlock.ScId() != scid || stateBlock.StateIndex() != stateIndex {
			log.Errorf("unexpected state tx data",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		// posting to the committee's queue
		sm.committee.ReceiveMessage(commtypes.StateTransactionMsg{
			Transaction: tx,
		})
	}()
}

// adding state update to the 'pending' map
func (sm *StateManager) addPendingStateUpdate(stateUpdate state.StateUpdate) bool {
	var varState state.VariableState
	if sm.solidVariableState != nil {
		if stateUpdate.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			// only interested in updates to the current state
			return false
		}
		varState = sm.solidVariableState.Apply(stateUpdate)
	} else {
		if stateUpdate.StateIndex() != 0 {
			// in the origin, only interested in updates with index 0
			return false
		}
		varState = state.CreateOriginVariableState(stateUpdate)
	}

	stateHash := hashing.GetHashValue(varState)
	existingRecord, ok := sm.pendingStateUpdates[stateHash]
	if ok && existingRecord.stateUpdate.StateTransactionId() != sctransaction.NilId {
		// corresponding pending update already exist
		return false
	}
	sm.pendingStateUpdates[stateHash] = &pendingStateUpdate{
		stateUpdate:       stateUpdate,
		nextVariableState: varState,
	}
	return true
}
