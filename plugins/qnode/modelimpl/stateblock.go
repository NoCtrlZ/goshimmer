package modelimpl

import (
	"errors"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

type mockStateBlock struct {
	assemblyId            *HashValue
	config                *mockConfig
	err                   error
	stateIndex            uint32
	stateChainOutputIndex uint16
	vars                  generic.ValueMap
	requestRef            *sc.RequestRef
}

type mockConfig struct {
	state        *mockStateBlock
	id           *HashValue
	vars         generic.ValueMap
	minReward    uint64
	ownersMargin byte
}

func newConfig(id *HashValue, state *mockStateBlock, minReward uint64, ownersMargin byte) *mockConfig {
	return &mockConfig{
		state:        state,
		id:           id,
		vars:         generic.NewFlatValueMap(),
		minReward:    minReward,
		ownersMargin: ownersMargin,
	}
}

func (cfg *mockConfig) Id() *HashValue {
	return cfg.id
}

func (cfg *mockConfig) Vars() generic.ValueMap {
	return cfg.vars
}

func (cfg *mockConfig) AssemblyAccount() *HashValue {
	addr, ok := cfg.vars.GetString(sc.MAP_KEY_ASSEMBLY_ACCOUNT)
	if !ok {
		return NilHash
	}
	ret, err := HashValueFromString(addr)
	if err != nil {
		return NilHash
	}
	return ret
}

func (cfg *mockConfig) OwnerAccount() *HashValue {
	addr, ok := cfg.vars.GetString(sc.MAP_KEY_OWNER_ACCOUNT)
	if !ok {
		return NilHash
	}
	ret, err := HashValueFromString(addr)
	if err != nil {
		return NilHash
	}
	return ret
}

func (cfg *mockConfig) MinimumReward() uint64 {
	return cfg.minReward
}

func (cfg *mockConfig) OwnersMargin() byte {
	return cfg.ownersMargin
}

func (cfg *mockConfig) With(config sc.Config) sc.Config {
	cfg.id = config.Id()
	cfg.ownersMargin = config.OwnersMargin()
	cfg.minReward = config.MinimumReward()
	cfg.vars = config.Vars().Clone()
	return cfg
}

func newStateBlock(aid, cid *HashValue, reqRef *sc.RequestRef) sc.State {
	ret := &mockStateBlock{
		assemblyId: aid,
		requestRef: reqRef,
		vars:       generic.NewFlatValueMap(),
	}
	ret.config = newConfig(cid, ret, 0, 0)
	return ret
}

// state

func (st *mockStateBlock) AssemblyId() *HashValue {
	return st.assemblyId
}

func (st *mockStateBlock) StateChainOutputIndex() uint16 {
	return st.stateChainOutputIndex
}

func (st *mockStateBlock) RequestRef() (*sc.RequestRef, bool) {
	if st.requestRef == nil {
		return nil, false
	}
	return st.requestRef, true
}

func (st *mockStateBlock) Vars() generic.ValueMap {
	return st.vars
}

func (st *mockStateBlock) StateIndex() uint32 {
	return st.stateIndex
}

func (st *mockStateBlock) Encode() generic.Encode {
	return st
}

func (st *mockStateBlock) Config() sc.Config {
	return st.config
}

func (st *mockStateBlock) Error() error {
	return st.err
}

func (st *mockStateBlock) WithError(err error) sc.State {
	st.err = err
	return st
}

func (st *mockStateBlock) WithStateIndex(idx uint32) sc.State {
	st.stateIndex = idx
	return st
}

func (st *mockStateBlock) WithVars(vars generic.ValueMap) sc.State {
	st.vars = vars.Clone()
	return st
}

func (st *mockStateBlock) WithStateChainOutputIndex(idx uint16) sc.State {
	st.stateChainOutputIndex = idx
	return st
}

// Encode
func (st *mockStateBlock) Write(w io.Writer) error {
	_, err := w.Write(st.assemblyId.Bytes())
	if err != nil {
		return err
	}
	_, err = w.Write(st.config.id.Bytes())
	if err != nil {
		return err
	}
	requestRefExists := st.requestRef != nil
	err = tools.WriteBoolByte(w, requestRefExists)
	if err != nil {
		return err
	}
	if requestRefExists {
		_, err = w.Write(st.requestRef.TxId().Bytes())
		if err != nil {
			return err
		}
		err = tools.WriteUint16(w, st.requestRef.Index())
		if err != nil {
			return err
		}
	}
	isError := st.err != nil
	err = tools.WriteBoolByte(w, isError)
	if err != nil {
		return err
	}
	if isError {
		return tools.WriteBytes16(w, []byte(st.err.Error()))
	}

	// nen error state

	err = tools.WriteUint32(w, st.stateIndex)
	if err != nil {
		return err
	}
	err = tools.WriteUint16(w, st.stateChainOutputIndex)
	if err != nil {
		return err
	}
	err = st.vars.Encode().Write(w)
	if err != nil {
		return err
	}

	// config
	err = tools.WriteUint64(w, st.config.minReward)
	if err != nil {
		return err
	}
	err = tools.WriteByte(w, st.config.ownersMargin)
	if err != nil {
		return err
	}
	err = st.config.vars.Encode().Write(w)
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
	var requestRefExist bool
	var reqRef *sc.RequestRef

	err = tools.ReadBoolByte(r, &requestRefExist)
	if err != nil {
		return err
	}
	if requestRefExist {
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
		reqRef = sc.NewRequestRefFromTxId(&requestTxId, requestBlockIndex)
	}
	var isError bool
	err = tools.ReadBoolByte(r, &isError)
	if err != nil {
		return err
	}
	if isError {
		errTxt, err := tools.ReadBytes16(r)
		if err != nil {
			return err
		}
		st.assemblyId = &assemblyId
		st.config.id = &configId
		st.err = errors.New(string(errTxt))
		st.config.vars = nil
		st.stateIndex = 0
		st.stateChainOutputIndex = 0
		st.vars = nil
		st.requestRef = nil
		return nil
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
	vars := generic.NewFlatValueMap()
	err = vars.Encode().Read(r)
	if err != nil {
		return err
	}
	// config
	var minReward uint64
	err = tools.ReadUint64(r, &minReward)
	if err != nil {
		return err
	}
	var ownersMargin byte
	ownersMargin, err = tools.ReadByte(r)
	if err != nil {
		return err
	}
	cfgVars := generic.NewFlatValueMap()
	err = cfgVars.Encode().Read(r)
	if err != nil {
		return err
	}

	st.assemblyId = &assemblyId
	st.config.id = &configId
	st.err = nil
	st.config.minReward = minReward
	st.config.ownersMargin = ownersMargin
	st.config.vars = cfgVars
	st.stateIndex = stateIndex
	st.stateChainOutputIndex = stateChainOutputIndex
	st.vars = vars
	st.requestRef = reqRef
	return nil
}
