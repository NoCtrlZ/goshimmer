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

func StateUpdateStorageKey(color balance.Color, stateIndex uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(stateUpdateDbPrefix))
	buf.Write(color.Bytes())
	buf.Write(util.Uint32To4Bytes(stateIndex))
	return buf.Bytes()
}

func VariableStateStorageKey(color balance.Color) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(variableStateDbPrefix))
	buf.Write(color.Bytes())
	return buf.Bytes()
}

// loads state update with the given index
func LoadStateUpdate(scid sctransaction.ScId, stateIndex uint32) (StateUpdate, error) {
	storageKey := StateUpdateStorageKey(scid.Color(), stateIndex)
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil || !exist {
		return nil, err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := NewStateUpdate(scid, stateIndex)
	if err = ret.Read(rdr); err != nil {
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
		Key:   StateUpdateStorageKey(su.essence.scid.Color(), su.essence.stateIndex),
		Value: hashing.MustBytes(su),
	})
}

// loads variable state from db
func LoadVariableState(scid sctransaction.ScId) (VariableState, error) {
	storageKey := VariableStateStorageKey(scid.Color())
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil || !exist {
		return nil, err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := &mockVariableState{
		scid:       scid,
		merkleHash: hashing.HashValue{},
	}
	if err = ret.Read(rdr); err != nil {
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
		Key:   VariableStateStorageKey(vs.scid.Color()),
		Value: hashing.MustBytes(vs),
	})
}
