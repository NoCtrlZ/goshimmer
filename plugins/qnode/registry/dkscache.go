package registry

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"sync"
)

var (
	dkscache      = make(map[hashing.HashValue]*tcrypto.DKShare)
	dkscacheMutex = &sync.Mutex{}
)

func CacheDKShare(dkshare *tcrypto.DKShare, id *hashing.HashValue) {
	dkscacheMutex.Lock()
	dkscache[*id] = dkshare
	dkscacheMutex.Unlock()
}

func UncacheDKShare(id *hashing.HashValue) {
	dkscacheMutex.Lock()
	delete(dkscache, *id)
	dkscacheMutex.Unlock()
}

func GetDKShare(id *hashing.HashValue) (*tcrypto.DKShare, bool, error) {
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()

	ret, ok := dkscache[*id]
	if ok {
		return ret, true, nil
	}
	var err error
	ok, err = ExistDKShareInRegistry(id)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := LoadDKShare(id, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[*id] = ks
	return ks, true, nil
}
