package modelimpl

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockUTXO struct {
	id        *hashing.HashValue
	inputSigs []generic.SignedBlock
	inputs    []value.Input
	outputs   []value.Output
}

func newUTXOTransfer() value.UTXOTransfer {
	return &mockUTXO{
		inputs:  make([]value.Input, 0),
		outputs: make([]value.Output, 0),
	}
}

func (tr *mockUTXO) ShortStr() string {
	balOut := uint64(0)
	for _, outp := range tr.outputs {
		balOut += outp.Value()
	}
	balIn := uint64(0)
	for _, inp := range tr.inputs {
		b, err := value.GetOutputAddrValue(inp.OutputRef())
		if err != nil {
			return err.Error()
		}
		balIn += b.Value
	}
	return fmt.Sprintf("numIn: %d numOut %d numSig: %d in: %d out: %d",
		len(tr.inputs), len(tr.outputs), len(tr.inputSigs), balIn, balOut)
}

func (tr *mockUTXO) Id() *hashing.HashValue {
	if tr.id != nil {
		return tr.id
	}
	if len(tr.inputs) == 0 && len(tr.outputs) == 0 {
		return hashing.NilHash
	}
	var buf bytes.Buffer
	for _, inp := range tr.inputs {
		_ = inp.Encode().Write(&buf)
	}
	for _, outp := range tr.outputs {
		_ = outp.Encode().Write(&buf)
	}
	// No signatures included !!!!
	tr.id = hashing.HashData(buf.Bytes())
	return tr.id
}

func (tr *mockUTXO) Inputs() []value.Input {
	return tr.inputs
}

func (tr *mockUTXO) Outputs() []value.Output {
	return tr.outputs
}

func (tr *mockUTXO) InputSignatures() ([]generic.SignedBlock, error) {
	if len(tr.inputSigs) != 0 {
		return tr.inputSigs, nil
	}
	outputsToSign, err := tr.collectOutputsToSign()
	if err != nil {
		return nil, err
	}
	ret := make([]generic.SignedBlock, 0, len(outputsToSign))
	for _, addr := range sortedAccounts(outputsToSign) {
		outs := outputsToSign[*addr]
		data := make([][]byte, 0, 2*len(outs))
		for _, o := range outs {
			data = append(data, o.TransferId.Bytes())
			data = append(data, tools.Uint16To2Bytes(o.OutputIndex))
		}
		ret = append(ret, generic.NewSignedBlock(addr.Clone(), hashing.HashData(data...)))
	}
	tr.inputSigs = ret
	return ret, nil
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
	if tr.id != nil {
		panic("AddInput: can't modify finalized transfer")
	}
	tr.inputs = append(tr.inputs, inp)
	tr.inputSigs = nil
	return uint16(len(tr.inputs) - 1)
}

func (tr *mockUTXO) AddOutput(outp value.Output) uint16 {
	if tr.id != nil {
		panic("AddInput: can't modify finalized transfer")
	}
	tr.outputs = append(tr.outputs, outp)
	return uint16(len(tr.outputs) - 1)
}

func (tr *mockUTXO) Encode() generic.Encode {
	return tr
}

func (tr *mockUTXO) collectOutputsToSign() (map[hashing.HashValue][]*generic.OutputRef, error) {
	ret := make(map[hashing.HashValue][]*generic.OutputRef)
	for _, inp := range tr.inputs {
		av, err := value.GetOutputAddrValue(inp.OutputRef())
		if err != nil {
			return nil, err
		}
		if _, ok := ret[*av.Addr]; !ok {
			ret[*av.Addr] = make([]*generic.OutputRef, 0)
		}
		ret[*av.Addr] = append(ret[*av.Addr], inp.OutputRef())
	}
	return ret, nil
}

func sortedAccounts(data map[hashing.HashValue][]*generic.OutputRef) []*hashing.HashValue {
	ret := make([]*hashing.HashValue, 0, len(data))
	for addr := range data {
		ret = append(ret, addr.Clone())
	}
	hashing.SortHashes(ret)
	return ret
}

// Encode

func (tr *mockUTXO) Write(w io.Writer) error {
	//if len(tr.inputs) == 0 || len(tr.outputs) == 0 || len(tr.inputSigs) == 0 {
	//	return fmt.Errorf("transfer not completed")
	//}
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
	err = generic.WriteSignedBlocks(w, tr.inputSigs)
	if err != nil {
		return err
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
	sigs, err := generic.ReadSignedBlocks(r)
	if err != nil {
		return err
	}

	tr.outputs = outps
	tr.inputs = inps
	tr.inputSigs = sigs
	return nil
}
