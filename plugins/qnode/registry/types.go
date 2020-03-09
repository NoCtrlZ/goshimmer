package registry

import . "github.com/iotaledger/goshimmer/plugins/qnode/hashing"

// general data about assembly
// each node in the assembly configurations must have identical assembly data

type AssemblyData struct {
	AssemblyId  *HashValue `json:"assembly_id"`
	OwnerPubKey *HashValue `json:"owner_pub_key"`
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
	accounts      map[HashValue]bool
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
