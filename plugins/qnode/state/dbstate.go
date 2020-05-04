package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"github.com/iotaledger/hive.go/database"
)

// database keys

const (
	stateUpdateDbPrefix      = "upd_"
	variableStateDbPrefix    = "vs_"
	requestProcessedDbPrefix = "rq_"
)

func stateUpdateStorageKey(color balance.Color, stateIndex uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(stateUpdateDbPrefix))
	buf.Write(color.Bytes())
	buf.Write(util.Uint32To4Bytes(stateIndex))
	return buf.Bytes()
}

func variableStateStorageKey(color balance.Color) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(variableStateDbPrefix))
	buf.Write(color.Bytes())
	return buf.Bytes()
}

func requestStorageKey(reqid *sctransaction.RequestId) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(requestProcessedDbPrefix))
	buf.Write(reqid.Bytes())
	return buf.Bytes()
}

// loads state update with the given index
func LoadStateUpdate(scid sctransaction.ScId, stateIndex uint32) (StateUpdate, error) {
	storageKey := stateUpdateStorageKey(scid.Color(), stateIndex)
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
	ret := NewStateUpdate(sctransaction.NilScId, 0)
	if err = ret.Read(rdr); err != nil {
		return nil, err
	}
	// check consistency of the stored object
	if ret.ScId() != scid || ret.StateIndex() != stateIndex {
		return nil, fmt.Errorf("LoadStateUpdate: invalid state update record in DB at state index %d scid %s",
			stateIndex, scid.String())
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
		Key:   stateUpdateStorageKey(su.scid.Color(), su.stateIndex),
		Value: hashing.MustBytes(su),
	})
}

// loads variable state from db
func LoadVariableState(scid sctransaction.ScId) (VariableState, error) {
	storageKey := variableStateStorageKey(scid.Color())
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
		Key:   variableStateStorageKey(vs.scid.Color()),
		Value: hashing.MustBytes(vs),
	})
}

// marks request processed
// TODO time when processed, cleanup the index after some time and so on
func MarkRequestProcessed(reqid *sctransaction.RequestId, errstr string) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   requestStorageKey(reqid),
		Value: []byte(errstr),
	})
}

// checks if request is processed
func IsRequestProcessed(reqid *sctransaction.RequestId) (bool, error) {
	storageKey := requestStorageKey(reqid)
	dbase, err := db.Get()
	if err != nil {
		return false, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil {
		return false, err
	}
	return exist, nil
}

// retrieves associated error string to the "request processed" record (if exists)
func RequestProcessedErrorString(reqid *sctransaction.RequestId) (string, error) {
	storageKey := requestStorageKey(reqid)
	dbase, err := db.Get()
	if err != nil {
		return "", err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return "", err
	}
	return string(entry.Value), nil
}
