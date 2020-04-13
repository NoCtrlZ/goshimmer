package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/mockclientlib"
	"io/ioutil"
	"os"
)

var (
	inCh   = make(chan mockclientlib.InMessage, 10)
	outCh  = make(chan []byte, 10)
	params *configParams
)

type configParams struct {
	WebAddress    string               `json:"web_address"`
	WebPort       int                  `json:"web_port"`
	MockPubAddr   string               `json:"mock_pub_address"`
	MockPubPort   int                  `json:"mock_pub_port"`
	SCDescription string               `json:"description"`
	N             uint16               `json:"n"`
	T             uint16               `json:"t"`
	Addresses     []*hashing.HashValue `json:"addresses"`
	Peers         []*registry.PortAddr `json:"peers"`
	ConfigId      *hashing.HashValue   `json:"config_id"`
	Scid          *hashing.HashValue   `json:"scid"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: mocknode <input file path>\n")
		os.Exit(1)
	}
	fname := os.Args[1]
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params = &configParams{}
	err = json.Unmarshal(data, params)
	if err != nil {
		panic(err)
	}
	if len(params.Peers) != int(params.N) || params.N < params.T || params.N < 4 {
		panic("wrong assembly size parameters or number of peers")
	}
	params.Scid = hashing.HashStrings(params.SCDescription)
	fmt.Printf("assembly dscr = %s\n", params.SCDescription)
	fmt.Printf("assembly id = %s\n", params.Scid.String())

	initGlobals()

	err = mockclientlib.RunPub(params.MockPubPort, outCh)
	if err != nil {
		panic(err)
	}
	fmt.Printf("publishing on port %d\n", params.MockPubPort)
	for _, pa := range params.Peers {
		go mockclientlib.ReadSub(fmt.Sprintf("tcp://%s:%d", pa.Addr, pa.Port), inCh)
	}
	go inLoop()
	runWebServer()
}

func inLoop() {
	for msg := range inCh {
		processInMsg(msg)
	}
}

func postTx(tx sc.Transaction) {
	var buf bytes.Buffer
	vtx, err := tx.ValueTx()
	if err != nil {
		fmt.Printf("postTx 1: %v\n", err)
		return
	}
	err = vtx.Encode().Write(&buf)
	if err != nil {
		fmt.Printf("postTx 2: %v\n", err)
		return
	}
	outCh <- buf.Bytes()
}

func processInMsg(msg mockclientlib.InMessage) {
	vtx, err := value.ParseTransaction(msg.Data)
	if err != nil {
		fmt.Printf("value.ParseTransaction: %v\n", err)
		return
	}
	tx, err := sc.ParseTransaction(vtx)
	if err != nil {
		fmt.Printf("sc.ParseTransaction: %v\n", err)
		return
	}

	fmt.Printf("processInMsg: tx id = %s trid = %s from %s\n",
		tx.Id().Short(), tx.Transfer().Id().Short(), msg.Uri)

	err = sc.VerifySignatures(tx, clientapi.NewDummyKeyPool())
	if err != nil {
		fmt.Printf("VerifySignedBlocks: %v\n", err)
		return
	}
	fmt.Printf("Signatures OK\n")
	if err := ldb.PutTransaction(vtx); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	setSCState(tx)
	// publish to nodes
	outCh <- msg.Data
}
