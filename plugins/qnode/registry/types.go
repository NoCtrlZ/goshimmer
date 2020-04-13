package registry

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

// general data about assembly
// each node in the assembly configurations must have identical assembly data
// TODO not final. Will change!!!!

type SCData struct {
	Scid        *HashValue `json:"scid"`
	OwnerPubKey *HashValue `json:"owner_pub_key"`
	Description string     `json:"description"`
	Program     string     `json:"program"`
}

// configuration of the assembly.
// can be several configurations of the assembly
// each node in the assembly configurations must have identical copy of each configuration (except index)

// changing configurations of the assembly means changing assembly account:
//  -- stopping the assembly
//  -- returning iotas for unprocessed requests (cancel pending requests)
//  -- setting new configuration in the chain
//  -- starting assembly again, with new configuration

type ConfigData struct {
	Index    uint16     `json:"index"`
	ConfigId *HashValue `json:"config_id,omitempty"`
	Created  int64      `json:"created"`

	// config id 1 (hash)  -- collection of keys + node IP locations
	// config id 0 (hash)  -- collection of keys
	Scid      *HashValue   `json:"scid"`
	N         uint16       `json:"n"`
	T         uint16       `json:"t"`
	Addresses []*HashValue `json:"addresses"` // addresses controlled by the smart contracts
	// -- end config id 0
	NodeLocations []*PortAddr `json:"node_locations"`
	// -- end of config id 1
	accounts map[HashValue]bool
}

type PortAddr struct {
	Port int    `json:"port"`
	Addr string `json:"addr"`
}

func (oa *PortAddr) AdjustedIP() (string, int) {
	if oa.Addr == "localhost" {
		return "127.0.0.1", oa.Port
	}
	return oa.Addr, oa.Port
}

func (oa *PortAddr) String() string {
	return fmt.Sprintf("%s:%d", oa.Addr, oa.Port)
}
