package fairroulette

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"sort"
)

func getRandom(signature string) uint32 {
	// calculate random uint64
	sigBin, err := hex.DecodeString(signature)
	if err != nil {
		return 0
	}
	h := hashing.HashData(sigBin)
	// rnd < pot, uniformly distributed [0,pot)
	return tools.Uint32From4Bytes(h.Bytes()[:4])
}

func runRoulette(lockedBets []*generic.OutputRefWithAddrValue, signature string) (*hashing.HashValue, uint64, error) {
	// assert signature != ""
	// sum up by payout address
	byPayout := make(map[hashing.HashValue]uint64)
	pot := uint64(0)
	for _, bet := range lockedBets {
		v, _ := byPayout[*bet.Addr]
		byPayout[*bet.Addr] = v + bet.Value
		pot += bet.Value
	}
	// to have fixed order on dictionary keys
	sortedAddresses := sortAddresses(byPayout)

	// calculate random uint64
	sigBin, err := hex.DecodeString(signature)
	if err != nil {
		return nil, 0, err
	}
	h := hashing.HashData(sigBin)
	// rnd < pot, uniformly distributed [0,pot)
	rnd := tools.Uint64From8Bytes(h.Bytes()[:8]) % pot

	// run roulette
	var runSum uint64
	for _, addr := range sortedAddresses {
		if runSum <= rnd && rnd <= runSum+byPayout[*addr] {
			return addr, pot, nil
		}
		runSum += byPayout[*addr]
	}
	return nil, 0, fmt.Errorf("runRoulette: inconsistency")
}

type arrToSort []*hashing.HashValue

func (s arrToSort) Len() int {
	return len(s)
}

func (s arrToSort) Less(i, j int) bool {
	return bytes.Compare(s[i].Bytes(), s[j].Bytes()) < 0
}

func (s arrToSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortAddresses(byAddr map[hashing.HashValue]uint64) []*hashing.HashValue {
	toSort := make([]*hashing.HashValue, 0, len(byAddr))
	for addr := range byAddr {
		toSort = append(toSort, addr.Clone())
	}
	sort.Sort(arrToSort(toSort))
	return toSort
}
