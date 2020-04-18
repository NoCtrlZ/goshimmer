package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
)

type ioParams struct {
	Hosts  []*registry.PortAddr `json:"hosts"`
	SCData registry.SCData      `json:"sc_data"`
}

type ioGetParams struct {
	Hosts []*registry.PortAddr `json:"hosts"`
	ScId  *hashing.HashValue   `json:"scid"`
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "deploy contract to iota",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("contract path is required\n")
						os.Exit(1)
					}
					fmt.Printf("Reading input from file: %s\n", c.Args().Get(0))
					Newsc(c.Args().Get(0))
					return nil
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Get deployed contract data",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("contract path is required\n")
						os.Exit(1)
					}
					fmt.Printf("Requesting SC data from nodes\n")
					fmt.Printf("Reading input from file: %s\n", c.Args().Get(0))
					GetSc(c.Args().Get(0))
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "Get deployed contract list",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("node url is required\n")
						os.Exit(1)
					}
					fmt.Printf("Requesting SC list from nodes\n")
					GetScList(c.Args().Get(0))
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

func Newsc(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	params.SCData.Scid = hashing.HashStrings(params.SCData.Description)
	params.SCData.OwnerPubKey = hashing.HashData(params.SCData.Scid.Bytes())
	params.SCData.Program = "dummy"
	for _, h := range params.Hosts {
		err = apilib.PutSCData(h.Addr, h.Port, &params.SCData)
		if err != nil {
			fmt.Printf("PutSCData: %v\n", err)
		} else {
			fmt.Printf("PutSCData success: %s:%d\n", h.Addr, h.Port)
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

func GetSc(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioGetParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("SC ID = %s\n", params.ScId.String())

	res := make(map[string]*registry.SCData)
	for _, h := range params.Hosts {
		scData, err := apilib.GetSCdata(h.Addr, h.Port, params.ScId)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		res[h.String()] = scData
		fmt.Printf("GetSCData from %s: success\n", h.String())
	}
	data, err = json.MarshalIndent(res, "", " ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = ioutil.WriteFile(fname+".get_resp.json", data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	if len(res) == 0 {
		fmt.Printf("no data was retrieved")
		return
	}
	if len(res) == 1 {
		fmt.Printf("1 SC data record was retrived")
		return
	}
	fmt.Printf("%d SC data records was retrived\nChecking for consistency...\n", len(res))
	// checking if all data records are identical
	var scDataCheck *registry.SCData
	var inconsistent bool
	for _, scData := range res {
		if scDataCheck == nil {
			scDataCheck = scData
			continue
		}
		if !scDataCheck.Scid.Equal(scData.Scid) {
			inconsistent = true
			break
		}
		if scDataCheck.Description != scData.Description {
			inconsistent = true
			break
		}
		if !scDataCheck.OwnerPubKey.Equal(scData.OwnerPubKey) {
			inconsistent = true
			break
		}
		if scDataCheck.Program != scData.Program {
			inconsistent = true
			break
		}
	}
	if inconsistent {
		fmt.Printf("Some data records are different: consistency check FAIL\n")
	} else {
		fmt.Printf("ALL data records are equal between each other: consistency check PASS\n")
	}
}

func GetScList(url string) {
	scList, err := apilib.GetSClist(url)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println(scList)
}
