package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto/tbdn"
	"github.com/pkg/errors"
	"time"
)

type resultCalculated struct {
	res            *runtimeContext
	rsHash         *HashValue
	masterDataHash *HashValue
	// processing stateTx
	pullSent         bool
	whenLastPullSent time.Time
	// finalization
	finalized     bool
	finalizedWhen time.Time
}

func (op *AssemblyOperator) SignBlock(sigblock generic.SignedBlock) error {
	keyData, err := op.getKeyData(sigblock.Account())
	if err != nil {
		return err
	}
	pk := keyData.(*tcrypto.DKShare)
	signature, err := pk.SignShare(sigblock.SignedHash().Bytes())
	if err != nil {
		return fmt.Errorf("signing error: `%v`", err)
	}
	sigblock.SetSignature(signature, generic.SIG_TYPE_BLS_SIGSHARE)
	return nil
}

func (op *AssemblyOperator) VerifySignature(blk generic.SignedBlock) error {
	signature, typ := blk.GetSignature()
	if typ != generic.SIG_TYPE_BLS_SIGSHARE {
		return errors.New("only BLS sig shares expected")
	}
	keyData, err := op.getKeyData(blk.Account())
	if err != nil {
		return err
	}
	dkshare, ok := keyData.(*tcrypto.DKShare)
	if !ok {
		return errors.New("wrong type of key data")
	}
	err = dkshare.VerifySigShare(blk.SignedHash().Bytes(), signature)
	if err != nil {
		return fmt.Errorf("validatePushMessage: %v", err)
	}
	return nil
}

func (op *AssemblyOperator) getKeyData(addr *HashValue) (interface{}, error) {
	if !op.cfgData.AccountIsDefined(addr) {
		return nil, fmt.Errorf("account id %s is undefined for this configuration", addr.Short())
	}
	ret, ok, err := registry.GetDKShare(op.assemblyId, addr)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("key not found: %s", addr.Short())
	}
	return ret, nil
}

func (op *AssemblyOperator) signResult(tx sc.Transaction) error {
	return sc.SignTransaction(tx, op)
}

// result blocks must be at least quorum -1
// must all be signed block
func (op *AssemblyOperator) aggregateBlocks(resultBlocks []generic.SignedBlock, ownResultBlock generic.SignedBlock) error {
	ownSigShare, _ := ownResultBlock.GetSignature()
	ownAddr := ownResultBlock.Account()
	ownDigest := ownResultBlock.SignedHash()

	tmp, err := op.getKeyData(ownAddr)
	if err != nil {
		return err
	}
	pkdata, ok := tmp.(*tcrypto.DKShare)
	if !ok {
		return errors.New("aggregateBlocks: unknown signature type")
	}
	sigShares := make([][]byte, 0, len(resultBlocks)+1)
	sigShares = append(sigShares, ownSigShare)

	for _, blk := range resultBlocks {
		if !blk.Account().Equal(ownAddr) {
			return errors.New("aggregateBlocks: signatures with different key ids can't be aggregated")
		}
		if !blk.SignedHash().Equal(ownDigest) {
			return errors.New("aggregateBlocks: signatures with different data digests can't be aggregated")
		}
		sigShare, _ := blk.GetSignature()
		sigShares = append(sigShares, sigShare)
	}
	recoveredSignature, err := tbdn.Recover(pkdata.Suite, pkdata.PubPoly,
		ownDigest.Bytes(), sigShares, int(op.requiredQuorum()), int(op.assemblySize()))
	if err != nil {
		return err
	}
	ownResultBlock.SetSignature(recoveredSignature, generic.SIG_TYPE_BLS_FINAL)
	pubKeyBin, err := pkdata.PubKeyMaster.MarshalBinary()
	if err != nil {
		return err
	}
	ownResultBlock.SetPublicKey(pubKeyBin)
	return nil
}
