package value

import "github.com/iotaledger/goshimmer/plugins/qnode/hashing"

func GetOutputAddrValue(inp Input) (*hashing.HashValue, uint64) {
	tr := inp.GetInputTransfer()
	if tr == nil {
		return hashing.NilHash, 1
	}
	output := tr.Outputs()[inp.OutputRef().OutputIndex()]
	return output.Address(), output.Value()
}
