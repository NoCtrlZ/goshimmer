package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/network/udp"
	"io/ioutil"
	"net"
	"os"
)

var (
	srv    *udp.UDPServer
	inCh   = make(chan *wrapped, 10)
	params *configParams
)

type wrapped struct {
	senderIndex uint16
	tx          sc.Transaction
}

type configParams struct {
	WebAddress          string               `json:"web_address"`
	WebPort             int                  `json:"web_port"`
	UDPAddress          string               `json:"udp_address"`
	UDPPort             int                  `json:"udp_port"`
	AssemblyDescription string               `json:"description"`
	N                   uint16               `json:"n"`
	T                   uint16               `json:"t"`
	Accounts            []*hashing.HashValue `json:"accounts"`
	Peers               []*registry.PortAddr `json:"peers"`
	ConfigId            *hashing.HashValue   `json:"config_id"`
	AssemblyId          *hashing.HashValue   `json:"assembly_id"`
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
	params.AssemblyId = hashing.HashStrings(params.AssemblyDescription)
	fmt.Printf("assembly dscr = %s\n", params.AssemblyDescription)
	fmt.Printf("assembly id = %s\n", params.AssemblyId.String())

	initGlobals()

	srv = udp.NewServer(parameters.UDP_BUFFER_SIZE)
	srv.Events.Start.Attach(events.NewClosure(func() {
		fmt.Printf("MockTangle ServerInstance started\n")
	}))
	srv.Events.ReceiveData.Attach(events.NewClosure(receiveUDPData))

	srv.Events.Error.Attach(events.NewClosure(func(err error) {
		fmt.Printf("Error: %v\n", err)
	}))

	go inLoop()

	fmt.Printf("listen UDP on %s:%d\n", params.UDPAddress, params.UDPPort)
	go srv.Listen(params.UDPAddress, params.UDPPort)

	runWebServer()
}

func receiveUDPData(updAddr *net.UDPAddr, data []byte) {
	idx := findSenderIndex(updAddr)
	fmt.Printf("---- received %d bytes\n", len(data))

	tx, err := decodeUDPMsg(data)
	if err != nil {
		fmt.Printf("decode tx error: %v\n", err)
		return
	}
	state, ok := tx.State()
	if !ok {
		fmt.Printf("RECEIVED NOT STATE UPDATE from %d\n", idx)
		return
	}

	if state.Error() == nil {
		fmt.Printf("RECEIVED from %d tx = %s stateIdx = %d\n", idx, tx.Id().Short(), state.StateIndex())
	} else {
		fmt.Printf("ERROR from %d tx = %s err = %v\n", idx, tx.Id().Short(), state.Error())
	}

	err = sc.VerifySignatures(tx, clientapi.NewDummyKeyPool())
	if err != nil {
		fmt.Printf("VerifySignedBlocks: %v\n", err)
		return
	}
	fmt.Printf("Signatures OK\n")
	vtx, _ := tx.ValueTx()
	if err := ldb.PutTransaction(vtx); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	setSCState(tx)

	postMsg(&wrapped{
		senderIndex: idx,
		tx:          tx,
	})
}

func findSenderIndex(updAddr *net.UDPAddr) uint16 {
	for i, a := range params.Peers {
		if updAddr.Port == a.Port && updAddr.IP.String() == a.Addr {
			return uint16(i)
		}
	}
	return qserver.MockTangleIdx
}

func inLoop() {
	for msg := range inCh {
		processMsg(msg)
	}
}

func postMsg(msg *wrapped) {
	inCh <- msg
}

func decodeUDPMsg(data []byte) (sc.Transaction, error) {
	_, _, _, msg, err := qserver.UnwrapUDPPacket(data)
	if err != nil {
		return nil, err
	}
	vtx, err := value.ParseTransaction(msg)
	if err != nil {
		return nil, err
	}

	tx, err := sc.ParseTransaction(vtx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func processMsg(msg *wrapped) {
	tx := msg.tx
	fmt.Printf("processMsg: tx id = %s trid = %s from sender %d\n",
		tx.Id().Short(), tx.Transfer().Id().Short(), msg.senderIndex)
	vtx, _ := tx.ValueTx()

	var buf bytes.Buffer
	err := vtx.Encode().Write(&buf)
	if err != nil {
		return
	}
	data := buf.Bytes()
	sentTo := sendToNodes(data)
	if sentTo != nil {
		fmt.Printf("sent to %+v\n", sentTo)
	}
}

func sendToNodes(data []byte) []string {
	wrapped := qserver.WrapUDPPacket(hashing.NilHash, qserver.MockTangleIdx, 0, data)
	if len(wrapped) > parameters.UDP_BUFFER_SIZE {
		fmt.Printf("sendToNodes: len(wrapped) > parameters.UDP_BUFFER_SIZE. Message wasn't sent")
		return nil
	}

	sentTo := make([]string, 0)
	for _, op := range params.Peers {
		addr := net.UDPAddr{
			IP:   net.ParseIP(op.Addr),
			Port: op.Port,
			Zone: "",
		}
		_, err := srv.GetSocket().WriteTo(wrapped, &addr)
		if err != nil {
			fmt.Printf("error while sending data to %+v: %v\n", addr, err)
		}
		sentTo = append(sentTo, addr.String())
	}
	return sentTo
}
