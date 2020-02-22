package generic

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func Bytes(b Encode) ([]byte, error) {
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Hash(e Encode) *hashing.HashValue {
	b, _ := Bytes(e)
	return hashing.HashData(b)
}

func NewOutputRef(trid *hashing.HashValue, outidx uint16) *OutputRef {
	return &OutputRef{
		trid:   trid,
		outIdx: outidx,
	}
}

func (or *OutputRef) TransferId() *hashing.HashValue {
	return or.trid
}

func (or *OutputRef) OutputIndex() uint16 {
	return or.outIdx
}

func (or *OutputRef) Id() *hashing.HashValue {
	var buf bytes.Buffer
	_, _ = buf.Write(or.trid.Bytes())
	_, _ = buf.Write(tools.Uint16To2Bytes(or.outIdx))
	return hashing.HashData(buf.Bytes())
}
