package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"time"
)

var webHostsAll = []*registry.PortAddr{
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

var nodeUDPAddressesAll = []*registry.PortAddr{
	{4000, "127.0.0.1"},
	{4001, "127.0.0.1"},
	{4002, "127.0.0.1"},
	{4003, "127.0.0.1"},
	{4004, "127.0.0.1"},
	{4005, "127.0.0.1"},
	{4006, "127.0.0.1"},
	{4007, "127.0.0.1"},
	{4008, "127.0.0.1"},
	{4009, "127.0.0.1"},
}

const (
	firstN              = 10
	assemblyDescription = "test assembly 2"
	N                   = uint16(10)
	T                   = uint16(7)
	cfgId1              = "d85036600bb75389dae0d501d983bbe0d1edb3251a5590816c314d9f390cb85f" // 1 account
	cfgId2              = "eddb2656a97ff6be411aac0d2fddb1fd1cc7de42905eaa742a09031ee921c261" // 2 accounts
)

var accStrings2 = []string{
	"c59de480c9ea21705b0d66299f14e9976308e3d7802971271b5eedd9e1f7a9ad",
	"158284bb4c1f33342681832bed2b807286744f098f7f1c58289169ba7b603415",
}

var accStrings1 = []string{
	"c59de480c9ea21705b0d66299f14e9976308e3d7802971271b5eedd9e1f7a9ad",
}

var accStringsN10 = []string{
	"ceb5579e21e651dd48c47eea42fa7e6ddd0732e3df9ef8de127d693b977ea4e1",
	"60ef310872f2b4d09cb2fa43e843b514fc21d3ea72b268a39d822b8ca9d5fd19",
}

var accStrings = accStringsN10
var webHosts = webHostsAll[:firstN]
var nodeUDPAddresses = nodeUDPAddressesAll[:firstN]

func main() {

	var err error
	assemblyId := hashing.HashStrings(assemblyDescription)
	accounts := make([]*hashing.HashValue, len(accStrings))
	for i, addr := range accStrings {
		accounts[i], err = hashing.HashValueFromString(addr)
		if err != nil {
			panic(err)
		}
	}
	cd := registry.ConfigData{
		Created:       time.Now().UnixNano(),
		AssemblyId:    assemblyId,
		N:             N,
		T:             T,
		NodeAddresses: nodeUDPAddresses,
		Accounts:      accounts,
	}
	for i, h := range webHosts {
		cd.Index = uint16(i)
		configId, err := apilib.NewConfiguration(h.Addr, h.Port, &cd)
		if err != nil {
			fmt.Printf("NewConfiguration: %v\n", err)
		} else {
			fmt.Printf("NewConfiguration: %s:%d config id = %s\n", h.Addr, h.Port, configId.String())
		}
	}
}
