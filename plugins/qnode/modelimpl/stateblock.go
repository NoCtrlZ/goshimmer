package modelimpl

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockStateBlock struct {
	assemblyId            *HashValue
	configId              *HashValue
	stateIndex            uint32
	stateChainOutputIndex uint16
	configVars            generic.ValueMap
	stateVars             generic.ValueMap
	requestTxId           *HashValue
	requestBlockIndex     uint16
}

func newStateBlock(aid, cid, reqId *HashValue, reqIdx uint16) sc.State {
	return &mockStateBlock{
		assemblyId:        aid,
		configId:          cid,
		requestTxId:       reqId,
		requestBlockIndex: reqIdx,
		configVars:        generic.NewFlatValueMap(),
		stateVars:         generic.NewFlatValueMap(),
	}
}

// state builder

func (st *mockStateBlock) InitFromPrev(prev sc.State) {
	st.stateIndex = prev.StateIndex() + 1
	st.stateVars = prev.StateVars()
	st.configVars = prev.ConfigVars()
}

func (st *mockStateBlock) SetConfigVars(vars generic.ValueMap) {
	st.configVars = vars
}

func (st *mockStateBlock) SetStateVars(vars generic.ValueMap) {
	st.stateVars = vars
}

func (st *mockStateBlock) SetRequest(reqTxId *HashValue, reqBlkIdx uint16) {
	st.requestTxId = reqTxId
	st.requestBlockIndex = reqBlkIdx
}

func (st *mockStateBlock) SetStateChainOutputIndex(idx uint16) {
	st.stateChainOutputIndex = idx
}

// state

func (st *mockStateBlock) Builder() sc.StateBuilder {
	return st
}

func (st *mockStateBlock) AssemblyId() *HashValue {
	return st.assemblyId
}

func (st *mockStateBlock) ConfigId() *HashValue {
	return st.configId
}

func (st *mockStateBlock) StateChainOutputIndex() uint16 {
	return st.stateChainOutputIndex
}

func (st *mockStateBlock) RequestRef() (*HashValue, uint16) {
	return st.requestTxId, st.requestBlockIndex
}

func (st *mockStateBlock) StateChainAddress() *HashValue {
	addr, ok := st.configVars.GetString("state_chain_addr")
	if !ok {
		return NilHash
	}
	ret, err := HashValueFromString(addr)
	if err != nil {
		return NilHash
	}
	return ret
}

func (st *mockStateBlock) RequestChainAddress() *HashValue {
	addr, ok := st.configVars.GetString("request_chain_addr")
	if !ok {
		return NilHash
	}
	ret, err := HashValueFromString(addr)
	if err != nil {
		return NilHash
	}
	return ret
}

func (st *mockStateBlock) OwnerChainAddress() *HashValue {
	addr, ok := st.configVars.GetString("owner_chain_addr")
	if !ok {
		return NilHash
	}
	ret, err := HashValueFromString(addr)
	if err != nil {
		return NilHash
	}
	return ret
}

func (st *mockStateBlock) ConfigVars() generic.ValueMap {
	return st.configVars
}

func (st *mockStateBlock) StateVars() generic.ValueMap {
	return st.stateVars
}

func (st *mockStateBlock) StateIndex() uint32 {
	return st.stateIndex
}

func (st *mockStateBlock) Encode() generic.Encode {
	return st
}

// Encode
func (st *mockStateBlock) Write(w io.Writer) error {
	_, err := w.Write(st.assemblyId.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(st.configId.Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteUint32(w, st.stateIndex)
	if err != nil {
		return err
	}
	err = tools.WriteUint16(w, st.stateChainOutputIndex)
	if err != nil {
		return err
	}
	err = st.configVars.Encode().Write(w)
	if err != nil {
		return err
	}
	err = st.stateVars.Encode().Write(w)
	if err != nil {
		return err
	}
	_, err = w.Write(st.requestTxId.Bytes())
	if err != nil {
		return err
	}
	err = tools.WriteUint16(w, st.requestBlockIndex)
	return err
}

func (st *mockStateBlock) Read(r io.Reader) error {
	var assemblyId HashValue
	_, err := r.Read(assemblyId.Bytes())
	if err != nil {
		return err
	}
	var configId HashValue
	_, err = r.Read(configId.Bytes())
	if err != nil {
		return err
	}
	var stateIndex uint32
	err = tools.ReadUint32(r, &stateIndex)
	if err != nil {
		return err
	}
	var stateChainOutputIndex uint16
	err = tools.ReadUint16(r, &stateChainOutputIndex)
	if err != nil {
		return err
	}
	cfgVars := generic.NewFlatValueMap()
	err = cfgVars.Encode().Read(r)
	if err != nil {
		return err
	}
	stateVars := generic.NewFlatValueMap()
	err = stateVars.Encode().Read(r)
	if err != nil {
		return err
	}
	var requestTxId HashValue
	_, err = r.Read(requestTxId.Bytes())
	if err != nil {
		return err
	}
	var requestBlockIndex uint16
	err = tools.ReadUint16(r, &requestBlockIndex)
	if err != nil {
		return err
	}
	st.assemblyId = &assemblyId
	st.configId = &configId
	st.stateIndex = stateIndex
	st.stateChainOutputIndex = stateChainOutputIndex
	st.configVars = cfgVars
	st.stateVars = stateVars
	st.requestTxId = &requestTxId
	st.requestBlockIndex = requestBlockIndex
	return nil
}
