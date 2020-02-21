package modelimpl

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockRequestBlock struct {
	assemblyId        *HashValue
	isConfigUpdateReq bool
	chainOutputIndex  uint16
	vars              generic.ValueMap
}

func newRequestBock(aid *HashValue, isConfig bool, chainOutputIndex uint16) sc.Request {
	return &mockRequestBlock{
		assemblyId:        aid,
		isConfigUpdateReq: isConfig,
		chainOutputIndex:  chainOutputIndex,
		vars:              generic.NewFlatValueMap(),
	}
}

func (req *mockRequestBlock) Encode() generic.Encode {
	return req
}

func (req *mockRequestBlock) AssemblyId() *HashValue {
	return req.assemblyId
}

func (req *mockRequestBlock) IsConfigUpdateReq() bool {
	return req.isConfigUpdateReq
}

func (req *mockRequestBlock) Vars() generic.ValueMap {
	return req.vars
}

func (req *mockRequestBlock) RequestChainOutputIndex() uint16 {
	return req.chainOutputIndex
}

// Encode

func (req *mockRequestBlock) Write(w io.Writer) error {
	_, err := w.Write(req.assemblyId.Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteBoolByte(w, req.isConfigUpdateReq)
	if err != nil {
		return err
	}
	err = tools.WriteUint16(w, req.chainOutputIndex)
	if err != nil {
		return err
	}
	err = req.vars.Encode().Write(w)
	return err
}

func (req *mockRequestBlock) Read(r io.Reader) error {
	var aid HashValue
	_, err := r.Read(aid.Bytes())
	if err != nil {
		return err
	}
	var isConfig bool
	err = tools.ReadBoolByte(r, &isConfig)
	if err != nil {
		return err
	}
	var chainOutputIndex uint16
	err = tools.ReadUint16(r, &chainOutputIndex)
	if err != nil {
		return err
	}
	vars := generic.NewFlatValueMap()
	err = vars.Encode().Read(r)
	if err != nil {
		return err
	}
	req.assemblyId = &aid
	req.isConfigUpdateReq = isConfig
	req.chainOutputIndex = chainOutputIndex
	req.vars = vars
	return nil
}
