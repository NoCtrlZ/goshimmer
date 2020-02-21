package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func AuthorizedForAddress(transfer value.UTXOTransfer, addr *hashing.HashValue) bool {
	for _, inp := range transfer.Inputs() {
		addr, _ := value.GetAddrValue(inp.OutputRef())
		if addr.Equal(addr) {
			return true
		}
	}
	return false
}
