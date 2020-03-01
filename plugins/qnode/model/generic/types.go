package generic

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"io"
)

type SignatureType byte

const (
	SIG_TYPE_UNDEF        = SignatureType(0)
	SIG_TYPE_BLS_SIGSHARE = SignatureType(1)
	SIG_TYPE_BLS_FINAL    = SignatureType(2)
	SIG_TYPE_MOCKED       = SignatureType(3)
)

type SignedBlock interface {
	SignedHash() *HashValue
	Account() *HashValue
	SetSignature(signature []byte, signatureType SignatureType)
	GetSignature() ([]byte, SignatureType)
	SetPublicKey([]byte)
	GetPublicKey() []byte
	Encode() Encode
}

type Encode interface {
	Write(w io.Writer) error
	Read(r io.Reader) error
}

type OutputRef struct {
	trid   *HashValue // transfer id
	outIdx uint16
}

type OutputRefWithValue struct {
	OutputRef
	Value uint64
}

type KeyPool interface {
	SignBlock(SignedBlock) error
	VerifySignature(SignedBlock) error
	GetKeyData(*HashValue) (interface{}, error)
}
