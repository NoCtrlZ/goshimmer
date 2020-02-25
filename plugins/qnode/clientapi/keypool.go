package clientapi

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/rogpeppe/go-internal/cache"
)

type mockKeysPool struct {
}

// for testing

func NewDummyKeyPool() generic.KeyPool {
	return &mockKeysPool{}
}

func (kp *mockKeysPool) SignBlock(sigBlk generic.SignedBlock) error {
	var sig [128]byte
	h := hashing.HashData(sigBlk.SignedHash().Bytes(), sigBlk.Account().Bytes())
	copy(sig[:], h.Bytes())
	sigBlk.SetSignature(sig[:], generic.SIG_TYPE_MOCKED)
	return nil
}

func (kp *mockKeysPool) VerifySignature(sigBlk generic.SignedBlock) error {
	sig, typ := sigBlk.GetSignature()
	if typ != generic.SIG_TYPE_MOCKED {
		return fmt.Errorf("mocked signatire expected")
	}
	h := hashing.HashData(sigBlk.SignedHash().Bytes(), sigBlk.Account().Bytes())
	if bytes.Compare(sig[:cache.HashSize], h.Bytes()) == 0 {
		return nil
	}
	return fmt.Errorf("invalid signature")
}

func (kp *mockKeysPool) GetKeyData(_ *hashing.HashValue) (interface{}, error) {
	return nil, nil
}
