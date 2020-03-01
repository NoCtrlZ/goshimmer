package sc

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func AuthorizedForAddress(tx Transaction, account *hashing.HashValue) bool {
	for _, inp := range tx.Transfer().Inputs() {
		oav := value.MustGetOutputAddrValue(inp.OutputRef())
		if oav.Addr.Equal(account) {
			return true
		}
	}
	return false
}

func SignTransaction(tx Transaction, keys generic.KeyPool) error {
	sigblocks := tx.Signatures()
	var err error
	for _, sigblk := range sigblocks {
		err = keys.SignBlock(sigblk)
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifySignedBlocks(sigBlocks []generic.SignedBlock, keys generic.KeyPool) error {
	for _, blk := range sigBlocks {
		if err := keys.VerifySignature(blk); err != nil {
			return err
		}
	}
	return nil
}
