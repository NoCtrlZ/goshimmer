package registry

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/pkg/errors"
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

func GetDKShare(aid, id *hashing.HashValue) (*tcrypto.DKShare, bool, error) {
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()

	ret, ok := dkscache[*id]
	if ok {
		if ret.AssemblyId.Equal(aid) {
			return ret, true, nil
		}
		return nil, true, errors.New("GetDKShare: wrong assembly id")
	}
	var err error
	ok, err = tcrypto.ExistDKShareInRegistry(aid, id)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := tcrypto.LoadDKShare(aid, id, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[*id] = ks
	return ks, true, nil
}
