package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

var hostsAll = []*registry.PortAddr{
	{9090, "127.0.0.1"},
	{9091, "127.0.0.1"},
	{9092, "127.0.0.1"},
	{9093, "127.0.0.1"},
	{9094, "127.0.0.1"},
	{9095, "127.0.0.1"},
	{9096, "127.0.0.1"},
	{9097, "127.0.0.1"},
	{9098, "127.0.0.1"},
	{9099, "127.0.0.1"},
}

var hosts []*registry.PortAddr

const firstN = 10

const assemblyDescription = "test assembly 2"

func main() {
	hosts := hostsAll[:firstN]
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
