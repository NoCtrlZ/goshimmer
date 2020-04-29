package state

import (
	"bytes"
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

type mockStateUpdate struct {
	scid       sctransaction.ScId
	stateIndex uint32
	stateTxId  valuetransaction.Id
}

// StateUpdate

func NewStateUpdate(scid sctransaction.ScId, stateIndex uint32) StateUpdate {
	return &mockStateUpdate{
		scid:       scid,
		stateIndex: stateIndex,
	}
}

func (su *mockStateUpdate) StateIndex() uint32 {
	return su.stateIndex
}

func (su *mockStateUpdate) StateTransactionId() valuetransaction.Id {
	return su.stateTxId
}

func (su *mockStateUpdate) SetStateTransactionId(vtxId valuetransaction.Id) {
	su.stateTxId = vtxId
}

func (su *mockStateUpdate) write(w io.Writer) error {
	_, _ = w.Write(su.stateTxId[:])
	return nil
}

func (su *mockStateUpdate) read(r io.Reader) error {
	var b valuetransaction.Id
	_, _ = r.Read(b[:])
	su.stateTxId = b
	return nil
}

func (su *mockStateUpdate) Bytes() []byte {
	var buf bytes.Buffer
	_ = su.write(&buf)
	return buf.Bytes()
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
	return NewMockVariableState(stateUpdate.StateIndex(), *merkleHash)
}

func CreateOriginVariableState(stateUpdate StateUpdate) VariableState {
	return VariableState(nil).Apply(stateUpdate)
}

func (vs *mockVariableState) write(w io.Writer) error {
	_, _ = w.Write(util.Uint32To4Bytes(vs.stateIndex))
	_, _ = w.Write(vs.merkleHash.Bytes())
	return nil
}

func (vs *mockVariableState) read(r io.Reader) error {
	_ = util.ReadUint32(r, &vs.stateIndex)
	_, _ = r.Read(vs.merkleHash.Bytes())
	return nil
}

func (vs *mockVariableState) Bytes() []byte {
	var buf bytes.Buffer
	_ = vs.write(&buf)
	return buf.Bytes()
}
