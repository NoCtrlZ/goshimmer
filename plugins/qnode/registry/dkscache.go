package registry

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"fmt"
	"sync"
)

var (
	dkscache      = map[hashing.HashValue]*tcrypto.DKShare{}
	dkscacheMutex = &sync.Mutex{}
)

func CacheDKShare(dkshare *tcrypto.DKShare, id *hashing.HashValue) {
	dkscacheMutex.Lock()
	dkscache[*id] = dkshare
	fmt.Printf("dks cache length -> %v\n", len(dkscache))
	for key := range dkscache {
		fmt.Println("this is key")
		fmt.Println(key)
	}
	defer dkscacheMutex.Unlock()
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
	ok, err = tcrypto.ExistDKShareInRegistry(id)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := tcrypto.LoadDKShare(id, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[*id] = ks
	return ks, true, nil
}

func GetAllDKShare() ([]*tcrypto.DKShare, bool) {
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()

	var dkslist tcrypto.DKShareList
	fmt.Printf("all dks length -> %v\n", len(dkscache))
	for key := range dkscache {
		fmt.Println(key)
		value, ok := dkscache[key]
		if !ok {
			return nil, false
		}
		dkslist = append(dkslist, value)
	}
	return dkslist, true
}

func CacheLength() {
	fmt.Printf("length -> %v\n", len(dkscache))
}
