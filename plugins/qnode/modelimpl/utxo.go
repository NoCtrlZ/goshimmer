package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockUTXO struct {
	inputSigs map[hashing.HashValue]generic.SignedBlock
	inputs    []value.Input
	outputs   []value.Output
}

func newUTXOTransfer() value.UTXOTransfer {
	return &mockUTXO{
		inputs:  make([]value.Input, 0),
		outputs: make([]value.Output, 0),
	}
}

func (tr *mockUTXO) Id() *hashing.HashValue {
	return hashing.NilHash // TODO
}

func (tr *mockUTXO) Inputs() []value.Input {
	return tr.inputs
}

func (tr *mockUTXO) Outputs() []value.Output {
	return tr.outputs
}

func (tr *mockUTXO) InputSignatures() map[hashing.HashValue]generic.SignedBlock {
	if tr.inputSigs != nil {
		return tr.inputSigs
	}
	outputsToSign := tr.collectOutputsToSign()
	ret := make(map[hashing.HashValue]generic.SignedBlock)
	for addr, outs := range outputsToSign {
		data := make([][]byte, 0, 2*len(outs))
		for _, o := range outs {
			data = append(data, o.TransferId().Bytes())
			data = append(data, tools.Uint16To2Bytes(o.OutputIndex()))
		}
		ret[addr] = NewSignedBlock(addr.Clone(), hashing.HashData(data...))
	}
	tr.inputSigs = ret
	return ret
}

// except signatures

func (tr *mockUTXO) DataHash() *hashing.HashValue {
	datas := make([][]byte, 0, len(tr.inputs)+len(tr.outputs))
	for _, inp := range tr.inputs {
		data, _ := generic.Bytes(inp.Encode())
		datas = append(datas, data)
	}
	for _, outp := range tr.outputs {
		data, _ := generic.Bytes(outp.Encode())
		datas = append(datas, data)
	}
	return hashing.HashData(datas...)
}

func (tr *mockUTXO) AddInput(inp value.Input) uint16 {
	tr.inputs = append(tr.inputs, inp)
	return uint16(len(tr.inputs) - 1)
}

func (tr *mockUTXO) AddOutput(outp value.Output) uint16 {
	tr.outputs = append(tr.outputs, outp)
	return uint16(len(tr.outputs) - 1)
}

func (tr *mockUTXO) Encode() generic.Encode {
	return tr
}

func (tr *mockUTXO) collectOutputsToSign() map[hashing.HashValue][]*generic.OutputRef {
	ret := make(map[hashing.HashValue][]*generic.OutputRef)
	for _, inp := range tr.inputs {
		addr, _ := value.GetOutputAddrValue(inp)
		if _, ok := ret[*addr]; !ok {
			ret[*addr] = make([]*generic.OutputRef, 0)
		}
		ret[*addr] = append(ret[*addr], inp.OutputRef())
	}
	return ret
}

func (tr *mockUTXO) sortedAddresses(data map[hashing.HashValue][]*generic.OutputRef) []*hashing.HashValue {
	ret := make([]*hashing.HashValue, 0, len(data))
	for addr := range data {
		ret = append(ret, addr.Clone())
	}
	hashing.SortHashes(ret)
	return ret
}

// Encode

func (tr *mockUTXO) Write(w io.Writer) error {
	err := tools.WriteUint16(w, uint16(len(tr.inputs)))
	if err != nil {
		return err
	}
	for _, inp := range tr.inputs {
		err = inp.Encode().Write(w)
		if err != nil {
			return err
		}
	}
	err = tools.WriteUint16(w, uint16(len(tr.outputs)))
	if err != nil {
		return err
	}
	for _, out := range tr.outputs {
		err = out.Encode().Write(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tr *mockUTXO) Read(r io.Reader) error {
	var numInp uint16
	err := tools.ReadUint16(r, &numInp)
	if err != nil {
		return err
	}
	inps := make([]value.Input, numInp)
	for i := range inps {
		inps[i] = &mockInput{}
		err = inps[i].Encode().Read(r)
		if err != nil {
			return err
		}
	}

	var numOutp uint16
	err = tools.ReadUint16(r, &numOutp)
	if err != nil {
		return err
	}
	outps := make([]value.Output, numOutp)
	for i := range outps {
		outps[i] = &mockOutput{}
		err = outps[i].Encode().Read(r)
		if err != nil {
			return err
		}
	}
	tr.outputs = outps
	tr.inputs = inps
	return nil
}
