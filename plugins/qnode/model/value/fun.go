package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

func GetAddrValue(or *generic.OutputRef) (*hashing.HashValue, uint64) {
	tr := GetTransfer(or.TransferId())
	if tr == nil {
		return hashing.NilHash, 0
	}
	output := tr.Outputs()[or.OutputIndex()]
	return output.Address(), output.Value()
}
