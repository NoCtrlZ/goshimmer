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
const (
	assemblyDescription = "test assembly 2"
)
const (
	N          = uint16(10)
	T          = uint16(7)
	numKeySets = 2
)

func main() {
	hosts := hostsAll[:firstN]
	if int(N) != len(hosts) {
		panic("wrong params")
	}
	aid := hashing.HashStrings(assemblyDescription)

	fmt.Printf("creating new distributed key set at nodes %+v\n", hosts)
	fmt.Printf("assembly dscr = %s\n", assemblyDescription)
	fmt.Printf("assembly id = %s\n", aid.String())

	for i := 0; i < numKeySets; i++ {
		account, err := apilib.GenerateNewDistributedKeySet(hosts, aid, N, T)
		if err == nil {
			fmt.Printf("generated new keys for account id %s\n", account.String())
		} else {
			fmt.Printf("error: %v\n", err)
		}
	}
}
