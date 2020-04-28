package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"sync"
)

// unique key for a smart contract is Color of its scid

var (
	scontracts      = make(map[balance.Color]committee.Committee)
	scontractsMutex = &sync.RWMutex{}
)

func loadAllSContracts(ownAddr *registry.PortAddr) (int, error) {
	scontractsMutex.Lock()
	defer scontractsMutex.Unlock()

	sclist, err := registry.GetSCDataList(ownAddr)
	if err != nil {
		return 0, err
	}
	num := 0
	for _, scdata := range sclist {
		scontracts[scdata.ScId.Color()] = committee.NewSyncManager(scdata)
		num++
	}
	return num, nil
}

// is the SC with the color processed by this node
func isColorProcessedByNode(color balance.Color) bool {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	_, ok := scontracts[color]
	return ok
}

func getSyncMgr(color balance.Color) committee.SyncManager {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	ret, _ := scontracts[color]
	return ret
}
