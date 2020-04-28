package state

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"github.com/iotaledger/hive.go/database"
)

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

// loads state update with the given index
func LoadStateUpdate(scid sctransaction.ScId, stateIndex uint32) (StateUpdate, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(StateUpdateDBKey(scid.Color(), stateIndex))
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := &mockStateUpdate{
		scid:       scid,
		stateIndex: stateIndex,
	}
	if err = ret.read(rdr); err != nil {
		return nil, err
	}
	return ret, nil
}

// saves state update to db
func (su *mockStateUpdate) SaveToDb() error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   StateUpdateDBKey(su.scid.Color(), su.stateIndex),
		Value: su.Bytes(),
	})
}

// loads variable state from db
func LoadVariableState(scid sctransaction.ScId) (VariableState, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(VariableStateDBKey(scid.Color()))
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := &mockVariableState{
		scid:       scid,
		merkleHash: hashing.HashValue{},
	}
	if err = ret.read(rdr); err != nil {
		return nil, err
	}
	return ret, nil
}

// saves variable state to db
func (vs *mockVariableState) SaveToDb() error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   VariableStateDBKey(vs.scid.Color()),
		Value: vs.Bytes(),
	})
}
