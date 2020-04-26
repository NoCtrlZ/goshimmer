package dispatcher

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"sync"
)

// unique key for a smart contract is Color of its scid

var (
	scontracts      = make(map[balance.Color]*scontract)
	scontractsMutex = &sync.RWMutex{}
)

type scontract struct {
	scData *registry.SCData
}

func loadAllSContracts(ownAddr *registry.PortAddr) (int, error) {
	scontractsMutex.Lock()
	defer scontractsMutex.Unlock()

	sclist, err := registry.GetSCDataList(ownAddr)
	if err != nil {
		return 0, err
	}
	num := 0
	for _, scdata := range sclist {
		addSContract(scdata)
		num++
	}
	return num, nil
}

func addSContract(scData *registry.SCData) {
	scontracts[scData.ScId.Color()] = &scontract{scData: scData}
}

// is the SC with the color processed by this node
func isColorProcessedByNode(color balance.Color) bool {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	_, ok := scontracts[color]
	return ok
}
