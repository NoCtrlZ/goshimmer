package statemgr

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"time"
)

func (sm *StateManager) takeAction() {
	sm.queryStateUpdateFromPeerIfNeeded()
	if sm.checkStateTransition() {
		return
	}
}

// if state is corrupted, will never synchronize
// TODO detect corrupted state and stop the crap
func (sm *StateManager) checkStateTransition() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state updates we locate the first one, consistent with the next state transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	pending, ok := sm.pendingStateUpdates[varStateHash]
	if !ok {
		return false
	}
	if err := sm.saveStateToDb(pending.nextVariableState, pending.stateUpdate); err != nil {
		log.Errorf("failed to save state #%d: %v", pending.stateUpdate.StateIndex(), err)
		return false
	}

	sm.solidVariableState = pending.nextVariableState
	sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), varStateHash.Bytes())
	sm.permutationIndex = sm.committee.Size() - 1
	return false
}

func (sm *StateManager) queryStateUpdateFromPeerIfNeeded() {
	if sm.solidVariableState == nil {
		return
	}
	if sm.largestEvidencedStateIndex <= sm.solidVariableState.StateIndex()+1 {
		return
	}
	if !sm.syncMessageDeadline.Before(time.Now()) {
		return
	}
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
	}
}

func (sm *StateManager) accountLargestStateIndex(idx uint32) {
	if idx > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = idx
	}
}

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
		sm.committee.ProcessMessage(commtypes.StateTransactionMsg{
			Transaction: tx,
		})
	}()
}
