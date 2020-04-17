package main

import (
	"os"
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/urfave/cli/v2"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

type ioParams struct {
	Hosts  []*registry.PortAddr `json:"hosts"`
	SCData registry.SCData      `json:"sc_data"`
}

type ioGetParams struct {
	Hosts  []*registry.PortAddr `json:"hosts"`
	SCId registry.SCId      `json:"sc_id"`
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name: "new",
				Aliases: []string{"n"},
				Usage: "deploy contract to iota",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("one arg is required\n")
						os.Exit(1)
					}
					fmt.Printf("Contract path is %s\n", c.Args().Get(0))
					Newsc(c.Args().Get(0))
					return nil
				},
			},
			{
				Name: "get",
				Aliases: []string{"g"},
				Usage: "Get deployed contract",
				Action: func(c *cli.Context) error {
					if c.Args().Get(0) == "" {
						fmt.Printf("one arg is required\n")
						os.Exit(1)
					}
					fmt.Printf("Contract path is %s\n", c.Args().Get(0))
					Getsc(c.Args().Get(0))
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
	fmt.Printf("%+v\n", params)
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

func Getsc(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioGetParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	params.SCId.Scid = hashing.HashStrings(params.SCId.Description)
	for _, h := range params.Hosts {
		res, err := apilib.GetSCdata(h.Addr, h.Port, &params.SCId)
		if err != nil {
			panic(err)
		}
		data, err = json.MarshalIndent(res, "", " ")
		err = ioutil.WriteFile(fname+".resp.json", data, 0644)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Printf("GetSCData success: %s:%d\n", h.Addr, h.Port)
		return
	}
}
