package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"sync"
)

// unique key for a smart contract is Color of its scid

var (
	scontracts      = make(map[balance.Color]*committee.Committee)
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
		if c, err := committee.NewCommittee(scdata); err == nil {
			scontracts[scdata.ScId.Color()] = c
			num++
		} else {
			log.Warn(err)
		}
	}
	return num, nil
}

func getCommittee(color balance.Color) *committee.Committee {
	scontractsMutex.RLock()
	defer scontractsMutex.RUnlock()

	ret, _ := scontracts[color]
	return ret
}
