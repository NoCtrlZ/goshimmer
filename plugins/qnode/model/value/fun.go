package value

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func GetOutputAddrValue(or *generic.OutputRef) (*generic.OutputRefWithAddrValue, error) {
	tr := GetTransfer(or.TransferId)
	if tr == nil {
		return nil, fmt.Errorf("can't find transfer %s", or.TransferId.Short())
	}
	var addr *hashing.HashValue
	var value uint64
	if int(or.OutputIndex) >= len(tr.Outputs()) {
		return nil, fmt.Errorf("output index out of bounds")
	}
	output := tr.Outputs()[or.OutputIndex]
	addr = output.Address()
	value = output.Value()

	return &generic.OutputRefWithAddrValue{
		OutputRef: *or,
		Addr:      addr,
		Value:     value,
	}, nil
}

func OutputCanBeChained(or *generic.OutputRef, chainAccount *hashing.HashValue) bool {
	tmp, err := GetOutputAddrValue(or)
	if err != nil {
		return false
	}
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
