package signedblock

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type signedBlock struct {
	sigType    generic.SignatureType
	signedHash *hashing.HashValue
	addr       *hashing.HashValue
	signature  []byte
	pubKey     []byte
}

func InitSignedBlockImplementation() {
	generic.SetSignedBlockConstructor(newSignedBlock)
}

func newSignedBlock(addr, data *hashing.HashValue) generic.SignedBlock {
	return &signedBlock{
		signedHash: data,
		addr:       addr,
	}
}

func (sb *signedBlock) SignedHash() *hashing.HashValue {
	return sb.signedHash
}

func (sb *signedBlock) Account() *hashing.HashValue {
	return sb.addr
}

func (sb *signedBlock) SetSignature(signature []byte, signatureType generic.SignatureType) {
	sb.signature = signature
	sb.sigType = signatureType
}

func (sb *signedBlock) GetSignature() ([]byte, generic.SignatureType) {
	return sb.signature, sb.sigType
}

func (sb *signedBlock) SetPublicKey(pk []byte) {
	sb.pubKey = pk
}

func (sb *signedBlock) GetPublicKey() []byte {
	return sb.pubKey
}

func (sb *signedBlock) Encode() generic.Encode {
	return sb
}

func (sb *signedBlock) Write(w io.Writer) error {
	err := tools.WriteByte(w, byte(sb.sigType))
	if err != nil {
		return err
	}
	_, err = w.Write(sb.addr.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(sb.signedHash.Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteBytes16(w, sb.signature)
	return err
}

func (sb *signedBlock) Read(r io.Reader) error {
	b, err := tools.ReadByte(r)
	if err != nil {
		return err
	}
	var addr hashing.HashValue
	_, err = r.Read(addr[:])
	if err != nil {
		return err
	}
	var signedHash hashing.HashValue
	_, err = r.Read(signedHash[:])
	if err != nil {
		return err
	}
	sig, err := tools.ReadBytes16(r)
	if err != nil {
		return err
	}
	sb.sigType = generic.SignatureType(b)
	sb.addr = &addr
	sb.signedHash = &signedHash
	sb.signature = sig
	return nil
}
