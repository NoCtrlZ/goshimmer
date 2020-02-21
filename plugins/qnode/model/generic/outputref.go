package generic

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

type OutputRef struct {
	trid   *hashing.HashValue
	outIdx uint16
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
