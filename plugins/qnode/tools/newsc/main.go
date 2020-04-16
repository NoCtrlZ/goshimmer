package main

import (
	"os"
	"fmt"
	"log"
    "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "Smart Contract CLI",
		Usage: "This cli sends transaction for iota smart contract",
		Action: func(c *cli.Context) error {
			fmt.Printf("Arg is %q", c.Args().Get(0))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
