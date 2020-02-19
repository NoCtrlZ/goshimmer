package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func GetInputAddress(input value.Input) *hashing.HashValue {
	tr := input.GetInputTransfer()
	if tr == nil {
		return hashing.NilHash
	}
	return tr.Outputs()[input.OutputRef().OutputIndex()].Address()
}

func AuthorizedForAddress(transfer value.UTXOTransfer, addr *hashing.HashValue) bool {
	for _, inp := range transfer.Inputs() {
		if GetInputAddress(inp).Equal(addr) {
			return true
		}
	}
	return false
}
