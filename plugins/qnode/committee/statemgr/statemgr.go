// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"time"
)

type StateManager struct {
	committee commtypes.Committee

	// state is corrupted, SC can't proceed
	isCorrupted    bool
	isSolidified   bool
	isSynchronized bool

	// pending state updates are state updates calculated by the VM
	// in servant mode.
	// are not linked with the state transaction yet
	pendingStateUpdates []state.StateUpdate

	// state transaction with state index next to the lastSolidStateTransaction
	// it may be nil (does not exist or not fetched yet
	nextStateTransaction *sctransaction.Transaction

	// last solid transaction obtained from the tangle by the reference from the
	// solid state update
	lastSolidStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	solidVariableState state.VariableState

	// last state update stored in the database. Obtained by the index stored in variable state
	// In case of origin state index is 0
	lastSolidStateUpdate state.StateUpdate

	// last state transaction received from the tangle
	// it may be not solidified yet in the SC ledger
	// if it coincides with the lastSolidStateTransaction, the state is in sync, otherwise it is not
	lastStateTransaction *sctransaction.Transaction

	// largest state index seen from other messages. If this index is more than 1 step ahead then
	// the solid one, state is not synced
	largestEvidencedStateIndex uint32

	// synchronization status. It is reset when state becomes synchronized
	permutationOfPeers  []uint16
	permutationIndex    uint16
	syncMessageDeadline time.Time
}

func NewStateManager(committee commtypes.Committee) *StateManager {
	return &StateManager{
		committee:           committee,
		pendingStateUpdates: make([]state.StateUpdate, 0),
	}
}

func (sm *StateManager) setSynchronized(yes bool) {
	sm.isSynchronized = yes
	if sm.isSolidified && sm.isSynchronized {
		sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), sm.lastStateTransaction.Id().Bytes())
		sm.permutationIndex = sm.committee.Size() - 1
	}
}

func (sm *StateManager) accountLargestStateIndex(idx uint32) {
	if idx > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = idx
	}
}

func (sm *StateManager) IsCorruptedState() bool {
	return sm.isCorrupted
}

func (sm *StateManager) IsSolidifiedState() bool {
	return sm.isSolidified
}

func (sm *StateManager) IsSynchronizedState() bool {
	return sm.isSynchronized
}
