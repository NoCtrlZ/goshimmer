package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/pkg/errors"
)

var dkscache = make(map[hashing.HashValue]*tcrypto.DKShare)

func GetDKShare(aid *hashing.HashValue, id *hashing.HashValue) (*tcrypto.DKShare, bool, error) {
	ret, ok := dkscache[*id]
	if ok {
		if *ret.AssemblyId == *aid {
			return ret, true, nil
		}
		return nil, true, errors.New("wrong assembly id")
	}
	var err error
	ok, err = tcrypto.ExistDKShare(aid, id)
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
