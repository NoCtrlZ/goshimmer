// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"time"
)

type StateManager struct {
	committee commtypes.Committee

	// pending state updates are state updates calculated by the VM
	// in servant mode.
	// are not linked with the state transaction yet
	pendingStateUpdates map[hashing.HashValue]*pendingStateUpdate

	// state transaction with state index next to the lastSolidStateTransaction
	// it may be nil (does not exist or not fetched yet
	nextStateTransaction    *sctransaction.Transaction
	stateTransactionArrived time.Time

	// last solid transaction obtained from the tangle by the reference from the
	// solid state update
	lastSolidStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	solidVariableState state.VariableState

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

type pendingStateUpdate struct {
	// state update, not validated yet
	stateUpdate state.StateUpdate
	// resulting variable state applied to the solidVariableState
	nextVariableState state.VariableState
}

func New(committee commtypes.Committee) *StateManager {
	ret := &StateManager{
		committee:           committee,
		pendingStateUpdates: make(map[hashing.HashValue]*pendingStateUpdate),
	}
	go func() {
		ret.initLoadState()
	}()
	return ret
}
