package main

import (
    "os"
    "github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "Smart Contract CLI"
	app.Usage = "This cli sends transaction for iota smart contract"
	app.Version = "0.0.1"

	app.Run(os.Args)
}
