package committee

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
)

type SyncManager interface {
	IsDismissed() bool
	ProcessStateMsg(StateMsg)
	ProcessRequestMsg(*RequestMsg)
}

type StateMsg *sctransaction.Transaction

type RequestMsg struct {
	*sctransaction.Transaction
	reqIndex uint16
}

type syncManager struct {
}

func NewSyncManager(scdata *registry.SCData) SyncManager {
	return nil
}

func (sm *syncManager) IsDismissed() bool {
	panic("implement me")
}

func (sm *syncManager) ProcessStateMsg(StateMsg) {
	panic("implement me")
}

func (sm *syncManager) ProcessRequestMsg(*RequestMsg) {
	panic("implement me")
}
