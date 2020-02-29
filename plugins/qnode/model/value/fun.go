package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func MustGetOutputAddrValue(or *generic.OutputRef) (*hashing.HashValue, uint64) {
	tr := GetTransfer(or.TransferId())
	if tr == nil {
		return hashing.RandomHash(nil), 1 // TODO for testing only
	}
	output := tr.Outputs()[or.OutputIndex()]
	return output.Address(), output.Value()
}

func OutputCanBeChained(or *generic.OutputRef, chainAccount *hashing.HashValue) bool {
	addr, val := MustGetOutputAddrValue(or)
	return val == 1 && addr.Equal(chainAccount)
}

func SumOutputsToAddress(transfer UTXOTransfer, addr *hashing.HashValue, except []uint16) uint64 {
	var ret uint64
	for i, outp := range transfer.Outputs() {
		if tools.Uint16InList(uint16(i), except) {
			continue
		}
		if outp.Address().Equal(addr) {
			ret += outp.Value()
		}
	}
	return ret
}

func AddChainLink(tr UTXOTransfer, outRef *generic.OutputRef, addr *hashing.HashValue) uint16 {
	tr.AddInput(NewInputFromOutputRef(outRef))
	return tr.AddOutput(NewOutput(addr, 1))
}
