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
	Hosts        []*registry.PortAddr  `json:"hosts"`
	AssemblyData registry.AssemblyData `json:"assembly_data"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage newassembly <input file path>\n")
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
	params.AssemblyData.AssemblyId = hashing.HashStrings(params.AssemblyData.Description)
	params.AssemblyData.OwnerPubKey = hashing.HashData(params.AssemblyData.AssemblyId.Bytes())
	params.AssemblyData.Program = "dummy"
	fmt.Printf("%+v\n", params)
	for _, h := range params.Hosts {
		err = apilib.PutAssemblyData(h.Addr, h.Port, &params.AssemblyData)
		if err != nil {
			fmt.Printf("PutAssemblyData: %v\n", err)
		} else {
			fmt.Printf("PutAssemblyData success: %s:%d\n", h.Addr, h.Port)
		}
	}
	data, err = json.MarshalIndent(&params, "", " ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = ioutil.WriteFile("resp."+fname, data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
