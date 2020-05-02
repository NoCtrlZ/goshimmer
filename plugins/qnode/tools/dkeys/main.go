package main

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/urfave/cli/v2"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"log"
	"os"
)

type ioParams struct {
	Hosts     []*registry.PortAddr `json:"hosts"`
	N         uint16               `json:"n"`
	T         uint16               `json:"t"`
	NumKeys   uint16               `json:"num_keys"`
	Addresses []string             `json:"addresses"`
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Generate dks for each committee node",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("config path is required\n")
						os.Exit(1)
					}
					fmt.Printf("Reading input from file: %s\n", c.Args().Get(0))
					NewDKS(c.Args().Get(0))
					return nil
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Get dks from each committee nodes",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("contract path is required\n")
						os.Exit(1)
					}
					fmt.Printf("Requesting SC data from nodes\n")
					fmt.Printf("Reading input from file: %s\n", c.Args().Get(0))
					CheckDKS(c.Args().Get(0))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func NewDKS(fname string) {
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

	params.Addresses = make([]string, 0, params.NumKeys)
	for i := 0; i < int(params.NumKeys); i++ {
		addr, err := apilib.GenerateNewDistributedKeySet(params.Hosts, params.N, params.T)
		if err == nil {
			params.Addresses = append(params.Addresses, addr.String())
			fmt.Printf("generated new keys. Address = %s\n", addr.String())
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

func CheckDKS(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	params.Addresses = make([]string, 0, params.NumKeys)
	dks, err := apilib.GetDistributedKey(params.Hosts, params.N, params.T)
	if err == nil {
		fmt.Printf("successful for getting dks\n")
	} else {
		fmt.Printf("error: %v\n", err)
	}
	data, err = json.MarshalIndent(&dks, "", " ")
	if err != nil {
		fmt.Printf("error: get dks%v\n", err)
		return
	}
	err = ioutil.WriteFile(fname+".all_dks.resp.json", data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("crosschecking. Reading public keys back\n")
	for _, addr := range params.Addresses {
		fmt.Printf("crosschecking. Address %s\n", addr)
		a, err := address.FromBase58(addr)
		if err != nil {
			fmt.Printf("%s --> %v\n", addr, err)
			continue
		}
		resps := apilib.GetPublicKeyInfo(params.Hosts, &a)
		for i, r := range resps {
			if r == nil || r.Address != addr || r.N != params.N || r.T != params.T || int(r.Index) != i {
				fmt.Printf("%s --> returned none or wrong values\n", params.Hosts[i])
			} else {
				fmt.Printf("%s --> master pub key: %s\n", params.Hosts[i], r.PubKeyMaster)
			}
		}
	}
}
