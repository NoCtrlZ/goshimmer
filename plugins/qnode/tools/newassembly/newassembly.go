package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

var hosts = []registry.PortAddr{
	{8080, "127.0.0.1"},
	{8081, "127.0.0.1"},
	{8082, "127.0.0.1"},
	{8083, "127.0.0.1"},
}

const assemblyDescription = "test assembly 1"

func main() {
	assemblyId := hashing.HashStrings(assemblyDescription)
	ownerPk := hashing.HashStrings("dummy").String() // for testing only

	od := registry.AssemblyData{
		AssemblyId:  assemblyId,
		OwnerPubKey: ownerPk,
		Description: assemblyDescription,
		Program:     "dummy",
	}
	fmt.Printf("%+v\n", od)
	var err error
	for _, h := range hosts {
		err = apilib.PutAssemblyData(h.Addr, h.Port, &od)
		if err != nil {
			fmt.Printf("PutAssemblyData: %v\n", err)
		} else {
			fmt.Printf("PutAssemblyData success: %s:%d\n", h.Addr, h.Port)
		}
	}
}
