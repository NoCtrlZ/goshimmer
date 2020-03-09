package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"io/ioutil"
	"os"
	"time"
)

type ioParams struct {
	Hosts               []*registry.PortAddr `json:"hosts"`
	AssemblyDescription string               `json:"description"`
	N                   uint16               `json:"n"`
	T                   uint16               `json:"t"`
	Accounts            []*hashing.HashValue `json:"accounts"`
	Peers               []*registry.PortAddr `json:"peers"`
	ConfigId            *hashing.HashValue   `json:"config_id"`
	AssemblyId          *hashing.HashValue   `json:"assembly_id"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage newdconfig <input file path>\n")
		os.Exit(1)
	}
	fname := os.Args[1]
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	if len(params.Hosts) != int(params.N) || params.N < params.T || params.N < 4 {
		panic("wrong assembly size parameters or number rof hosts")
	}
	params.AssemblyId = hashing.HashStrings(params.AssemblyDescription)
	fmt.Printf("assembly dscr = %s\n", params.AssemblyDescription)
	fmt.Printf("assembly id = %s\n", params.AssemblyId.String())

	cd := registry.ConfigData{
		Created:       time.Now().UnixNano(),
		AssemblyId:    params.AssemblyId,
		N:             params.N,
		T:             params.T,
		NodeAddresses: params.Peers,
		Accounts:      params.Accounts,
	}
	var configId *hashing.HashValue
	var wrongIds bool
	for i, h := range params.Hosts {
		cd.Index = uint16(i)
		_configId, err := apilib.NewConfiguration(h.Addr, h.Port, &cd)
		if err != nil {
			fmt.Printf("NewConfiguration: %v\n", err)
		} else {
			fmt.Printf("NewConfiguration: %s:%d config id = %s\n", h.Addr, h.Port, _configId.String())
		}
		if configId != nil && !configId.Equal(_configId) {
			fmt.Printf("error: nut equal configuration Ids returned")
			wrongIds = true
		}
		if configId == nil {
			configId = _configId
		}
	}
	if wrongIds {
		fmt.Printf("error occured")
		os.Exit(1)
	}
	params.ConfigId = configId
	data, err = json.MarshalIndent(&params, "", " ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = ioutil.WriteFile(fname+".resp.json", data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
