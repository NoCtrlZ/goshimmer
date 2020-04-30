package statemgr

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee/commtypes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"time"
)

func (sm *StateManager) takeAction() {
	if sm.isSolidified && !sm.isSynchronized {
		if sm.syncMessageDeadline.Before(time.Now()) {
			sm.sendSyncRequestToPeer()
		}
	}
}

func (sm *StateManager) sendSyncRequestToPeer() {
	sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
	data := hashing.MustBytes(&commtypes.GetStateUpdateMsg{
		StateIndex: sm.lastSolidStateUpdate.StateIndex() + 1,
	})
	for i := uint16(0); i < sm.committee.Size(); i++ {
		targetPeerIndex := sm.permutationOfPeers[sm.permutationIndex]
		if err := sm.committee.SendMsg(targetPeerIndex, commtypes.MsgGetStateUpdate, data); err == nil {
			break
		}
		sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
	}
}
