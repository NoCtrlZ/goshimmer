package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/pkg/errors"
)

// implements KeyPool interface

func (op *AssemblyOperator) SignBlock(sigblock generic.SignedBlock) error {
	keyData, err := op.GetKeyData(sigblock.Account())
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
	keyData, err := op.GetKeyData(blk.Account())
	if err != nil {
		return err
	}
	dkshare, ok := keyData.(*tcrypto.DKShare)
	if !ok {
		return errors.New("wrong type of key data: BLS expected")
	}
	switch typ {
	case generic.SIG_TYPE_BLS_SIGSHARE:
		err = dkshare.VerifySigShare(blk.SignedHash().Bytes(), signature)
	case generic.SIG_TYPE_BLS_FINAL:
		err = dkshare.VerifyMasterSignature(blk.SignedHash().Bytes(), signature)
	default:
		return errors.New("only BLS signatures expected")
	}
	if err != nil {
		return fmt.Errorf("ValidateSignature: %v", err)
	}
	return nil
}

func (op *AssemblyOperator) GetKeyData(addr *HashValue) (interface{}, error) {
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
