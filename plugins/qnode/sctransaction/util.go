package sctransaction

import (
	"crypto/rand"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
)

// make scid witg given address and random color
// ONLY FOR TESTING
func RandomScId(addr address.Address) *ScId {
	return NewScId(addr, RandomColor())
}

func RandomColor() *balance.Color {
	var ret balance.Color
	rndBytes := make([]byte, balance.ColorLength)

	if _, err := rand.Read(rndBytes); err != nil {
		panic(err)
	}
	copy(ret[:], rndBytes)
	return &ret
}

func SumBalancesOfColor(balances []*balance.Balance, color balance.Color) int64 {
	var ret int64
	for _, bal := range balances {
		if bal.Color() == color {
			ret += bal.Value()
		}
	}
	return ret
}
