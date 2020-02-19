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

func (iid *OutputRef) TransferId() *hashing.HashValue {
	return iid.trid
}

func (iid *OutputRef) OutputIndex() uint16 {
	return iid.outIdx
}
