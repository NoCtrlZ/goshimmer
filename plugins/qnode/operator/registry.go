package operator

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"time"
)

// general data about assembly
// each node in the assembly configurations must have identical assembly data

type AssemblyData struct {
	AssemblyId  *HashValue `json:"assembly_id"`
	Modified    int64      `json:"modified"`
	OwnerPubKey string     `json:"owner_pub_key"`
	Description string     `json:"description"`
	Program     string     `json:"program"`
}

// configuration of the assembly.
// can be several configurations of the assembly
// each node in the assembly configurations must have identical copy of each configuration

type ConfigData struct {
	Index             uint16       `json:"index"`
	ConfigId          *HashValue   `json:"config_id"`
	AssemblyId        *HashValue   `json:"assembly_id"`
	Created           int64        `json:"created"`
	OperatorAddresses []*PortAddr  `json:"addresses"`
	DKeyIds           []*HashValue `json:"dkey_ids"`
}

type PortAddr struct {
	Port int    `json:"port"`
	Addr string `json:"addr"`
}

func (oa PortAddr) AdjustedIP() (string, int) {
	if oa.Addr == "localhost" {
		return "127.0.0.1", oa.Port
	}
	return oa.Addr, oa.Port
}

func (ad *AssemblyData) Save() error {
	ad.Modified = time.Now().UnixNano()
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(ad)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   nil,
		Value: jsonData,
	})
}

func dbOpdataGroupKey() []byte {
	return []byte("opdata")
}

func LoadAllOperatorData() (map[HashValue]*AssemblyData, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	ret := make(map[HashValue]*AssemblyData)
	err = dbase.ForEachPrefix(dbOpdataGroupKey(), func(entry database.Entry) bool {
		opdata := &AssemblyData{}
		if err = json.Unmarshal(entry.Value, opdata); err == nil {
			ret[*opdata.AssemblyId] = opdata
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func dbCfgKey(aid, cid *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("cfgdata")
	buf.Write(aid.Bytes())
	buf.Write(cid.Bytes())
	return buf.Bytes()
}

func (pcfg *ConfigData) Save() error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(pcfg)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbCfgKey(pcfg.AssemblyId, pcfg.ConfigId),
		Value: jsonData,
	})
}

func ExistPrivateConfig(aid, cid *HashValue) (bool, error) {
	dbase, err := db.Get()
	if err != nil {
		return false, err
	}
	return dbase.Contains(dbCfgKey(aid, cid))
}

func LoadConfig(aid, cid *HashValue) (*ConfigData, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(dbCfgKey(aid, cid))
	if err != nil {
		return nil, err
	}
	ret := &ConfigData{}
	err = json.Unmarshal(entry.Value, ret)
	if err != nil {
		return nil, err
	}
	if !aid.Equal(ret.AssemblyId) {
		return nil, errors.New("inconsistency in configuration data")
	}
	return ret, nil
}
