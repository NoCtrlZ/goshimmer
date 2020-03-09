package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"io/ioutil"
	"os"
)

type ioParams struct {
	Hosts               []*registry.PortAddr `json:"hosts"`
	AssemblyDescription string               `json:"description"`
	N                   uint16               `json:"n"`
	T                   uint16               `json:"t"`
	NumKeys             uint16               `json:"num_keys"`
	Accounts            []*hashing.HashValue `json:"accounts"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage newdks <input file path>\n")
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
	aid := hashing.HashStrings(params.AssemblyDescription)
	fmt.Printf("assembly dscr = %s\n", params.AssemblyDescription)
	fmt.Printf("assembly id = %s\n", aid.String())

	params.Accounts = make([]*hashing.HashValue, 0, params.NumKeys)
	for i := 0; i < int(params.NumKeys); i++ {
		account, err := apilib.GenerateNewDistributedKeySet(params.Hosts, aid, params.N, params.T)
		params.Accounts = append(params.Accounts, account)
		if err == nil {
			fmt.Printf("generated new keys for account id %s\n", account.String())
		} else {
			fmt.Printf("error: %v\n", err)
		}
	}
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
