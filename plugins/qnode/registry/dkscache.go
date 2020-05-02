package registry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"sync"
)

var (
	dkscache      = make(map[address.Address]*tcrypto.DKShare)
	dkscacheMutex = &sync.RWMutex{}
)

// GetDKShare retrieves distributed key share from registry or the cache
// returns dkshare, exists flag and error
func GetDKShare(addr address.Address) (*tcrypto.DKShare, bool, error) {
	dkscacheMutex.RLock()
	ret, ok := dkscache[addr]
	if ok {
		defer dkscacheMutex.RUnlock()
		return ret, true, nil
	}
	// switching to write lock
	dkscacheMutex.RUnlock()
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()

	var err error
	ok, err = ExistDKShareInRegistry(addr)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := LoadDKShare(addr, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[addr] = ks
	return ks, true, nil
}
