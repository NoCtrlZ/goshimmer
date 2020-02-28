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
	vars              generic.ValueMap
}

func (req *mockRequestBlock) WithOutputIndices(chainIdx, rewardIdx, depositIdx uint16) sc.Request {
	req.vars.SetInt(sc.MAP_KEY_CHAIN_OUT_INDEX, int(chainIdx))
	req.vars.SetInt(sc.MAP_KEY_REWARD_OUT_INDEX, int(chainIdx))
	req.vars.SetInt(sc.MAP_KEY_DEPOSIT_OUT_INDEX, int(chainIdx))
	return req
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

func (req *mockRequestBlock) OutputIndices() (uint16, uint16, uint16) {
	chain, _ := req.vars.GetInt(sc.MAP_KEY_CHAIN_OUT_INDEX)
	reward, _ := req.vars.GetInt(sc.MAP_KEY_REWARD_OUT_INDEX)
	deposit, _ := req.vars.GetInt(sc.MAP_KEY_DEPOSIT_OUT_INDEX)
	return uint16(chain), uint16(reward), uint16(deposit)
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
	vars := generic.NewFlatValueMap()
	err = vars.Encode().Read(r)
	if err != nil {
		return err
	}
	req.assemblyId = &aid
	req.isConfigUpdateReq = isConfig
	req.vars = vars
	return nil
}
