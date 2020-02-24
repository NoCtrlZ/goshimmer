package main

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/modelimpl"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/signedblock"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/txdb"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/network/udp"
	"net"
)

const (
	address = "127.0.0.1"
	port    = 1000
)

var operators = []*registry.PortAddr{
	{4000, "127.0.0.1"},
	{4001, "127.0.0.1"},
	{4002, "127.0.0.1"},
	{4003, "127.0.0.1"},
}

var (
	srv  *udp.UDPServer
	inCh = make(chan *wrapped, 10)
	ldb  value.DB
)

type wrapped struct {
	senderIndex uint16
	tx          sc.Transaction
}

func main() {
	modelimpl.InitModelImplementation()
	signedblock.InitSignedBlockImplementation()

	srv = udp.NewServer(2048)
	srv.Events.Start.Attach(events.NewClosure(func() {
		fmt.Printf("MockTangle ServerInstance started\n")
	}))
	srv.Events.ReceiveData.Attach(events.NewClosure(receiveUDPData))

	srv.Events.Error.Attach(events.NewClosure(func(err error) {
		fmt.Printf("Error: %v\n", err)
	}))
	ldb = txdb.NewLocalDb()
	value.SetValuetxDB(ldb)

	go inLoop()

	fmt.Printf("listen UDP on %s:%d\n", address, port)
	go srv.Listen(address, port)

	runWebServer()
}

func receiveUDPData(updAddr *net.UDPAddr, data []byte) {
	idx := findSenderIndex(updAddr)
	msg, err := decodeUDPMsg(data)
	if err != nil {
		fmt.Printf("decode tx error: %v\n", err)
		return
	}
	postMsg(&wrapped{
		senderIndex: idx,
		tx:          msg,
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
	fmt.Printf("process sc tx: id = %s from sender %d\n", tx.Id().Short(), msg.senderIndex)

	vtx, _ := tx.ValueTx()
	if err := ldb.PutTransaction(vtx); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

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
