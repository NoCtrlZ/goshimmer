package generic

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto/tbdn"
)

func AggregateBLSBlocks(blocksToAggregate []SignedBlock, setBlock SignedBlock, keyPool KeyPool) error {
	if len(blocksToAggregate) < 1 {
		return fmt.Errorf("AggregateBLSBlocks: nothing to aggregate")
	}
	addr := blocksToAggregate[0].Account()
	digest := blocksToAggregate[0].SignedHash()

	tmp, err := keyPool.GetKeyData(addr)
	if err != nil {
		return err
	}
	pkdata, ok := tmp.(*tcrypto.DKShare)
	if !ok {
		return fmt.Errorf("AggregateBLSBlocks:BLS key type expected")
	}
	sigShares := make([][]byte, 0, len(blocksToAggregate))

	for _, blk := range blocksToAggregate {
		if !blk.Account().Equal(addr) {
			return fmt.Errorf("AggregateBLSBlocks: signatures with different key ids can't be aggregated")
		}
		if !blk.SignedHash().Equal(digest) {
			return fmt.Errorf("AggregateBLSBlocks: signatures with different data digests can't be aggregated")
		}
		sigShare, sigType := blk.GetSignature()
		if sigType != SIG_TYPE_BLS_SIGSHARE {
			return fmt.Errorf("AggregateBLSBlocks: unexpected signature type")
		}
		sigShares = append(sigShares, sigShare)
	}

	recoveredSignature, err := tbdn.Recover(pkdata.Suite, pkdata.PubPoly,
		digest.Bytes(), sigShares, int(pkdata.T), int(pkdata.N))
	if err != nil {
		return err
	}
	setBlock.SetSignature(recoveredSignature, SIG_TYPE_BLS_FINAL)
	pubKeyBin, err := pkdata.PubKeyMaster.MarshalBinary()
	if err != nil {
		return err
	}
	setBlock.SetPublicKey(pubKeyBin)
	return nil
}
