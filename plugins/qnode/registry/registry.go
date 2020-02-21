package registry

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/pkg/errors"
	"io"
)

// general data about assembly
// each node in the assembly configurations must have identical assembly data

type AssemblyData struct {
	AssemblyId  *HashValue `json:"assembly_id"`
	OwnerPubKey string     `json:"owner_pub_key"`
	Description string     `json:"description"`
	Program     string     `json:"program"`
}

// configuration of the assembly.
// can be several configurations of the assembly
// each node in the assembly configurations must have identical copy of each configuration (except index)

type ConfigData struct {
	Index    uint16     `json:"index"`
	ConfigId *HashValue `json:"config_id,omitempty"`
	Created  int64      `json:"created"`
	// config id
	AssemblyId    *HashValue   `json:"assembly_id"`
	N             uint16       `json:"n"`
	T             uint16       `json:"t"`
	NodeAddresses []*PortAddr  `json:"port_addr"`
	Accounts      []*HashValue `json:"accounts"`
	// TODO include additional vars
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

func dbOpdataGroupKey() []byte {
	return []byte("opdata")
}

func dbOpdateKey(aid *HashValue) []byte {
	var buf bytes.Buffer
	buf.Write(dbOpdataGroupKey())
	buf.Write(aid.Bytes())
	return buf.Bytes()
}

func (ad *AssemblyData) Save() error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(ad)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbOpdateKey(ad.AssemblyId),
		Value: jsonData,
	})
}

func dbCfgKey(aid, cid *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("cfgdata")
	buf.Write(aid.Bytes())
	buf.Write(cid.Bytes())
	return buf.Bytes()
}

func SaveConfig(cfg *ConfigData) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = cfg.Write(&buf)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbCfgKey(cfg.AssemblyId, cfg.ConfigId),
		Value: buf.Bytes(),
	})
}

func ExistConfig(aid, cid *HashValue) (bool, error) {
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
	err = ret.Read(bytes.NewReader(entry.Value))
	if err != nil {
		return nil, err
	}
	if !aid.Equal(ret.AssemblyId) {
		return nil, errors.New("inconsistency in configuration data")
	}
	return ret, nil
}

func (cfg *ConfigData) Write(w io.Writer) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return tools.WriteBytes32(w, data)
}

func (cfg *ConfigData) Read(r io.Reader) error {
	data, err := tools.ReadBytes32(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &cfg)
}

func ConfigId(cfg *ConfigData) *HashValue {
	var buf bytes.Buffer
	buf.Write(cfg.AssemblyId.Bytes())
	_ = tools.WriteUint16(&buf, cfg.N)
	_ = tools.WriteUint16(&buf, cfg.T)
	for _, na := range cfg.NodeAddresses {
		buf.WriteString(na.Addr)
		_ = tools.WriteUint32(&buf, uint32(na.Port))
	}
	for _, addr := range cfg.Accounts {
		buf.Write(addr.Bytes())
	}

	return HashData(buf.Bytes())
}
