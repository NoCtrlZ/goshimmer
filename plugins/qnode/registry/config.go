package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/pkg/errors"
	"io"
)

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
		Key:   dbCfgKey(cfg.Scid, cfg.ConfigId),
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
	if !aid.Equal(ret.Scid) {
		return nil, errors.New("inconsistency in configuration data")
	}
	ret.accounts = make(map[HashValue]bool)
	for _, addr := range ret.Addresses {
		ret.accounts[*addr] = true
	}
	return ret, nil
}

func (cfg *ConfigData) AccountIsDefined(addr *HashValue) bool {
	ok, _ := cfg.accounts[*addr]
	return ok
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
	buf.Write(cfg.Scid.Bytes())
	_ = tools.WriteUint16(&buf, cfg.N)
	_ = tools.WriteUint16(&buf, cfg.T)
	for _, na := range cfg.NodeLocations {
		buf.WriteString(na.Addr)
		_ = tools.WriteUint32(&buf, uint32(na.Port))
	}
	for _, addr := range cfg.Addresses {
		buf.Write(addr.Bytes())
	}

	return HashData(buf.Bytes())
}

func ValidateConfig(cfg *ConfigData) error {
	if len(cfg.Addresses) == 0 {
		return fmt.Errorf("0 accounts found")
	}
	if cfg.N < 4 {
		return fmt.Errorf("assembly size must be at least 4")
	}
	if cfg.T < cfg.N/2+1 {
		return fmt.Errorf("assembly quorum must be at least N/2+1 (2*N/3+1 recommended)")
	}
	if len(cfg.NodeLocations) != int(cfg.N) {
		return fmt.Errorf("number of nodes must be equal to the size of assembly N")
	}

	ok, err := ExistConfig(cfg.Scid, cfg.ConfigId)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("duplicated configuration id %s", cfg.ConfigId.Short())
	}
	// check consistency with account keys

	for _, addr := range cfg.Addresses {
		ks, ok, err := GetDKShare(addr)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("can't find account id %s", addr.Short())
		}
		if ks.N != cfg.N || ks.T != cfg.T {
			return fmt.Errorf("inconsistent size parameters with account id %s", addr.Short())
		}
	}

	if !differentAddresses(cfg.NodeLocations) {
		return fmt.Errorf("addresses of operator nodes must all be different")
	}
	return nil
}

func differentAddresses(addrs []*PortAddr) bool {
	if len(addrs) <= 1 {
		return true
	}
	for i := 0; i < len(addrs)-1; i++ {
		for j := i + 1; j < len(addrs); j++ {
			if addrs[i].Addr == addrs[j].Addr && addrs[i].Port == addrs[j].Port {
				return false
			}
		}
	}
	return true
}
