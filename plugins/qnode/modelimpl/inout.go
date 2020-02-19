package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockInput struct {
	oid generic.OutputRef
	tr  value.UTXOTransfer
}

func newInput(trid *hashing.HashValue, outidx uint16) value.Input {
	return &mockInput{
		oid: *generic.NewOutputRef(trid, outidx),
		tr:  nil}
}

func newOutput(addr *hashing.HashValue, value uint64) value.Output {
	return &mockOutput{
		addr:  addr,
		value: value,
	}
}

// Input

func (or *mockInput) OutputRef() *generic.OutputRef {
	return &or.oid
}

func (or *mockInput) GetInputTransfer() value.UTXOTransfer {
	return value.GetTransfer(or.oid.TransferId())
}

func (or *mockInput) Encode() generic.Encode {
	return or
}

type mockOutput struct {
	addr  *hashing.HashValue
	value uint64
}

// Encode

func (or *mockInput) Write(w io.Writer) error {
	_, err := w.Write(or.oid.TransferId().Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteUint16(w, or.oid.OutputIndex())
	return err
}

func (or *mockInput) Read(r io.Reader) error {
	var txid hashing.HashValue
	_, err := r.Read(txid.Bytes())
	if err != nil {
		return err
	}
	var idx uint16
	err = tools.ReadUint16(r, &idx)
	if err != nil {
		return err
	}
	or.oid = *generic.NewOutputRef(&txid, idx)
	return nil
}

// Output

func (out *mockOutput) Address() *hashing.HashValue {
	return out.addr
}

func (out *mockOutput) Value() uint64 {
	return out.value
}

func (out *mockOutput) Encode() generic.Encode {
	return out
}

// Encode
func (out *mockOutput) Write(w io.Writer) error {
	_, err := w.Write(out.addr.Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteUint64(w, out.value)
	if err != nil {
		return err
	}
	return err
}

func (out *mockOutput) Read(r io.Reader) error {
	var addr hashing.HashValue
	_, err := r.Read(addr.Bytes())
	if err != nil {
		return err
	}
	var value uint64
	err = tools.ReadUint64(r, &value)
	if err != nil {
		return err
	}
	out.addr = &addr
	out.value = value
	return nil
}
