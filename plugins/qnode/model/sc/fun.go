package sc

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func AuthorizedForAddress(tx Transaction, account *hashing.HashValue) bool {
	for _, inp := range tx.Transfer().Inputs() {
		addr, _ := value.GetAddrValue(inp.OutputRef())
		if addr.Equal(account) {
			return true
		}
	}
	return false
}
