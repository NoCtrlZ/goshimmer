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

func newRequestBock(aid *HashValue, isConfig bool) sc.Request {
	return &mockRequestBlock{
		assemblyId:        aid,
		isConfigUpdateReq: isConfig,
		vars:              generic.NewFlatValueMap(),
	}
}

func (req *mockRequestBlock) WithRequestChainOutputIndex(idx uint16) sc.Request {
	req.vars.SetInt(sc.MAP_KEY_CHAIN_OUT_INDEX, int(idx))
	return req
}

func (req *mockRequestBlock) WithRewardOutputIndex(idx uint16) sc.Request {
	req.vars.SetInt(sc.MAP_KEY_REWARD_OUT_INDEX, int(idx))
	return req
}

func (req *mockRequestBlock) WithDepositOutputIndex(idx uint16) sc.Request {
	req.vars.SetInt(sc.MAP_KEY_DEPOSIT_OUT_INDEX, int(idx))
	return req
}

func (req *mockRequestBlock) WithVars(vars generic.ValueMap) sc.Request {
	req.vars = vars
	return req
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

// 0 chain
// 1 reward
// 2 deposit

func (req *mockRequestBlock) MainOutputs(tx sc.Transaction) [3]*generic.OutputRefWithAddrValue {
	var ret [3]*generic.OutputRefWithAddrValue
	if chainOutIdx, ok := req.vars.GetInt(sc.MAP_KEY_CHAIN_OUT_INDEX); ok {
		ret[0] = &generic.OutputRefWithAddrValue{
			OutputRef: *generic.NewOutputRef(tx.Transfer().Id(), uint16(chainOutIdx)),
			Value:     tx.Transfer().Outputs()[chainOutIdx].Value(),
			Addr:      tx.Transfer().Outputs()[chainOutIdx].Address(),
		}
	}
	if rewardOutIdx, ok := req.vars.GetInt(sc.MAP_KEY_REWARD_OUT_INDEX); ok {
		ret[0] = &generic.OutputRefWithAddrValue{
			OutputRef: *generic.NewOutputRef(tx.Transfer().Id(), uint16(rewardOutIdx)),
			Value:     tx.Transfer().Outputs()[rewardOutIdx].Value(),
			Addr:      tx.Transfer().Outputs()[rewardOutIdx].Address(),
		}
	}
	if depositOutIdx, ok := req.vars.GetInt(sc.MAP_KEY_DEPOSIT_OUT_INDEX); ok {
		ret[0] = &generic.OutputRefWithAddrValue{
			OutputRef: *generic.NewOutputRef(tx.Transfer().Id(), uint16(depositOutIdx)),
			Value:     tx.Transfer().Outputs()[depositOutIdx].Value(),
			Addr:      tx.Transfer().Outputs()[depositOutIdx].Address(),
		}
	}
	return ret
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
