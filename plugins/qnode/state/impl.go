package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

// this is fake implementation of the VariableState and StateUpdate
// intended for testing
// VariableState is just a hashed value and stateIndex (not even needed for testing).
// State transition by applying state update to Variable state with Apply function
// is just a hashing of the previous VariableState
// the state update is empty always. The whole information about the state update is contained
// in the key: color and state index

type mockVariableState struct {
	scid       sctransaction.ScId
	stateIndex uint32
	merkleHash hashing.HashValue
}

type mockStateUpdateEssence struct {
	scid       sctransaction.ScId // persist in key
	stateIndex uint32             // persist in key
}

type mockStateUpdate struct {
	essence   *mockStateUpdateEssence
	stateTxId valuetransaction.Id
}

// StateUpdate

func NewStateUpdate(scid sctransaction.ScId, stateIndex uint32) StateUpdate {
	return &mockStateUpdate{
		essence: &mockStateUpdateEssence{
			scid:       scid,
			stateIndex: stateIndex,
		},
	}
}

// StateUpdateEssence

func (se *mockStateUpdateEssence) StateIndex() uint32 {
	return se.stateIndex
}

func (se *mockStateUpdateEssence) Write(w io.Writer) error {
	_, err := w.Write(hashing.NilHash.Bytes()) // dummy writing
	return err
}

func (se *mockStateUpdateEssence) Read(r io.Reader) error {
	// dummy reading
	var dummy hashing.HashValue
	_, err := r.Read(dummy.Bytes())
	return err
}

// StateUpdate interface

func (su *mockStateUpdate) Essence() StateUpdateEssence {
	return su.essence
}

func (su *mockStateUpdate) StateTransactionId() valuetransaction.Id {
	return su.stateTxId
}

func (su *mockStateUpdate) SetStateTransactionId(vtxId valuetransaction.Id) {
	su.stateTxId = vtxId
}

func (su *mockStateUpdate) IsAnchored() bool {
	return su.stateTxId != valuetransaction.Id(*hashing.NilHash)
}

func (su *mockStateUpdate) Write(w io.Writer) error {
	if err := su.essence.Write(w); err != nil {
		return err
	}
	_, err := w.Write(su.stateTxId[:])
	return err
}

func (su *mockStateUpdate) Read(r io.Reader) error {
	essence := &mockStateUpdateEssence{}
	if err := essence.Read(r); err != nil {
		return err
	}
	_, err := r.Read(su.stateTxId[:])
	return err
}

// VariableState

func NewMockVariableState(stateIndex uint32, hash hashing.HashValue) VariableState {
	return &mockVariableState{
		stateIndex: stateIndex,
		merkleHash: hash,
	}
}

func (vs *mockVariableState) StateIndex() uint32 {
	return vs.stateIndex
}

func (vs *mockVariableState) Apply(stateUpdate StateUpdate) VariableState {
	merkleHash := hashing.NilHash
	if vs != nil {
		merkleHash = hashing.HashData(vs.merkleHash.Bytes())
	}
	return NewMockVariableState(stateUpdate.Essence().StateIndex(), *merkleHash)
}

func CreateOriginVariableState(stateUpdate StateUpdate) VariableState {
	return VariableState(nil).Apply(stateUpdate)
}

func (vs *mockVariableState) Write(w io.Writer) error {
	if _, err := w.Write(util.Uint32To4Bytes(vs.stateIndex)); err != nil {
		return err
	}
	_, err := w.Write(vs.merkleHash.Bytes())
	return err
}

func (vs *mockVariableState) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &vs.stateIndex); err != nil {
		return err
	}
	_, err := r.Read(vs.merkleHash.Bytes())
	return err
}
