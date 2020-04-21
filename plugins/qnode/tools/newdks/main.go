package main

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
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
	Addresses []*hashing.HashValue `json:"addresses"`
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
					GetDKS(c.Args().Get(0))
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

	params.Addresses = make([]*hashing.HashValue, 0, params.NumKeys)
	for i := 0; i < int(params.NumKeys); i++ {
		addr, err := apilib.GenerateNewDistributedKeySet(params.Hosts, params.N, params.T)
		params.Addresses = append(params.Addresses, addr)
		if err == nil {
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

func GetDKS(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		fmt.Println("get dks")
		panic(err)
	}

	params.Addresses = make([]*hashing.HashValue, 0, params.NumKeys)
	dks, err := apilib.GetDistributedKey(params.Hosts, params.N, params.T)
	if err == nil {
		fmt.Printf("get dks\n")
	} else {
		fmt.Printf("get dks\n")
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
}
