package modelimpl

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"io"
)

type mockRequestBlock struct {
	assemblyId        *HashValue
	isConfigUpdateReq bool
	vars              generic.ValueMap
	chainOutputIndex  uint16
}

func newRequestBock(aid *HashValue, isConfig bool) sc.Request {
	return &mockRequestBlock{
		assemblyId:        aid,
		isConfigUpdateReq: isConfig,
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
	panic("implement me")
}

func (req *mockRequestBlock) Read(r io.Reader) error {
	panic("implement me")
}
