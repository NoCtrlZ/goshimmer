package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"time"
)

var hosts = []*registry.PortAddr{
	{8080, "127.0.0.1"},
	{8081, "127.0.0.1"},
	{8082, "127.0.0.1"},
	{8083, "127.0.0.1"},
}

const (
	assemblyDescription = "test assembly 1"
)

func main() {
	assemblyId := hashing.HashStrings(assemblyDescription)

	cd := registry.ConfigData{
		AssemblyId:        assemblyId,
		ConfigId:          hashing.NilHash,
		Created:           time.Now().UnixNano(),
		OperatorAddresses: hosts,
		Accounts:          []*hashing.HashValue{},
	}
	var err error
	for i, h := range hosts {
		cd.Index = uint16(i)
		err = apilib.NewConfiguration(h.Addr, h.Port, &cd)
		if err != nil {
			fmt.Printf("NewConfiguration: %v\n", err)
		} else {
			fmt.Printf("NewConfiguration: %s:%d\n", h.Addr, h.Port)
		}
	}
}
