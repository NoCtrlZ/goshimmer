package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

func GetAddrValue(or *generic.OutputRef) (*hashing.HashValue, uint64) {
	tr := GetTransfer(or.TransferId())
	if tr == nil {
		return hashing.RandomHash(nil), 1 // for testing only
	}
	output := tr.Outputs()[or.OutputIndex()]
	return output.Address(), output.Value()
}

func OutputCanBeChained(or *generic.OutputRef, chainAccount *hashing.HashValue) bool {
	addr, val := GetAddrValue(or)
	return val == 1 && addr.Equal(chainAccount)
}
