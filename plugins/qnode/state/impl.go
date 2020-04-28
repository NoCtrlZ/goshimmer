package state

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"github.com/iotaledger/hive.go/database"
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
	stateIndex uint32
	merkleHash hashing.HashValue
}

type mockStateUpdate struct {
	stateTxId valuetransaction.Id
}

// database keys

const (
	stateUpdateDbPrefix   = "upd_"
	variableStateDbPrefix = "vs_"
)

func StateUpdateDBKey(color balance.Color, stateIndex uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(stateUpdateDbPrefix))
	buf.Write(color.Bytes())
	buf.Write(util.Uint32To4Bytes(stateIndex))
	return buf.Bytes()
}

func VariableStateDBKey(color balance.Color) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(variableStateDbPrefix))
	buf.Write(color.Bytes())
	return buf.Bytes()
}

// StateUpdate

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

func (su *mockStateUpdate) LoadFromDb(color balance.Color, stateIndex uint32) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	entry, err := dbase.Get(StateUpdateDBKey(color, stateIndex))
	if err != nil {
		return err
	}
	rdr := bytes.NewReader(entry.Value)
	return su.read(rdr)
}

func (su *mockStateUpdate) SaveToDb(color balance.Color, stateIndex uint32) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   StateUpdateDBKey(color, stateIndex),
		Value: su.Bytes(),
	})
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
	merkleHash := hashing.HashData(vs.merkleHash.Bytes())
	return NewMockVariableState(vs.stateIndex+1, *merkleHash)
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

func (vs *mockVariableState) LoadFromDb(color balance.Color) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	entry, err := dbase.Get(VariableStateDBKey(color))
	if err != nil {
		return err
	}
	rdr := bytes.NewReader(entry.Value)
	return vs.read(rdr)
}

func (vs *mockVariableState) SaveToDb(color balance.Color) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   VariableStateDBKey(color),
		Value: vs.Bytes(),
	})
}
