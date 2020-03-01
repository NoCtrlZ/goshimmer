package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func MustGetOutputAddrValue(or *generic.OutputRef) *generic.OutputRefWithAddrValue {
	tr := GetTransfer(or.TransferId)
	var addr *hashing.HashValue
	var value uint64
	if tr == nil {
		addr = hashing.RandomHash(nil)
		value = 1 // TODO for testing only
	} else {
		output := tr.Outputs()[or.OutputIndex]
		addr = output.Address()
		value = output.Value()
	}

	return &generic.OutputRefWithAddrValue{
		OutputRef: *or,
		Addr:      addr,
		Value:     value,
	}
}

func OutputCanBeChained(or *generic.OutputRef, chainAccount *hashing.HashValue) bool {
	tmp := MustGetOutputAddrValue(or)
	return tmp.Value == 1 && tmp.Addr.Equal(chainAccount)
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
