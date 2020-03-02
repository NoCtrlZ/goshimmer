package main

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/clientapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/network/udp"
	"net"
)

const (
	address = "127.0.0.1"
	port    = 1000
	firstN  = 4
)

var operatorsAll = []*registry.PortAddr{
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

var (
	operators = operatorsAll[:firstN]
	srv       *udp.UDPServer
	inCh      = make(chan *wrapped, 10)
)

type wrapped struct {
	senderIndex uint16
	tx          sc.Transaction
}

func main() {
	initGlobals()

	srv = udp.NewServer(2048)
	srv.Events.Start.Attach(events.NewClosure(func() {
		fmt.Printf("MockTangle ServerInstance started\n")
	}))
	srv.Events.ReceiveData.Attach(events.NewClosure(receiveUDPData))

	srv.Events.Error.Attach(events.NewClosure(func(err error) {
		fmt.Printf("Error: %v\n", err)
	}))

	go inLoop()

	fmt.Printf("listen UDP on %s:%d\n", address, port)
	go srv.Listen(address, port)

	runWebServer()
}

func receiveUDPData(updAddr *net.UDPAddr, data []byte) {
	idx := findSenderIndex(updAddr)
	tx, err := decodeUDPMsg(data)
	if err != nil {
		fmt.Printf("decode tx error: %v\n", err)
		return
	}
	state, ok := tx.State()
	if !ok {
		fmt.Printf("RECEIVED NOT STATE UPDATEfrom %d\n", idx)
		return
	}

	if state.Error() == nil {
		fmt.Printf("RECEIVED from %d tx = %s stateIdx = %d\n", idx, tx.Id().Short(), state.StateIndex())
	} else {
		fmt.Printf("ERROR from %d tx = %s err = %v\n", idx, tx.Id().Short(), state.Error())
	}

	err = sc.VerifySignedBlocks(tx.Signatures(), clientapi.NewDummyKeyPool())
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

	postMsg(&wrapped{
		senderIndex: idx,
		tx:          tx,
	})
}

func findSenderIndex(updAddr *net.UDPAddr) uint16 {
	for i, a := range operators {
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
	sentTo := sendToNodes(buf.Bytes())
	fmt.Printf("sent to %+v\n", sentTo)
}

func sendToNodes(data []byte) []int16 {
	wrapped := qserver.WrapUDPPacket(hashing.NilHash, qserver.MockTangleIdx, 0, data)
	sentTo := make([]int16, 0)
	for idx, op := range operators {
		addr := net.UDPAddr{
			IP:   net.ParseIP(op.Addr),
			Port: op.Port,
			Zone: "",
		}
		_, err := srv.GetSocket().WriteTo(wrapped, &addr)
		if err != nil {
			fmt.Printf("error while sending data to %+v: %v\n", addr, err)
		}
		sentTo = append(sentTo, int16(idx))
	}
	return sentTo
}
