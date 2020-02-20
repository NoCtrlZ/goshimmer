package main

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
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

var srv *udp.Server
var inCh = make(chan interface{}, 10)

type wrapped struct {
	senderIndex int
	msg         interface{}
}

func main() {
	srv = udp.NewServer(2048)
	srv.Events.Start.Attach(events.NewClosure(func() {
		tools.Logf(0, "MockTangle ServerInstance started")
	}))
	srv.Events.Shutdown.Attach(events.NewClosure(func() {
		tools.Logf(0, "ServerInstance shutdown event")
	}))
	srv.Events.ReceiveData.Attach(events.NewClosure(receiveData))

	srv.Events.Error.Attach(events.NewClosure(func(err error) {
		tools.Logf(0, "Error: %v\n", err)
	}))
	go inLoop()

	tools.Logf(0, "listen UDP on %s:%d", address, port)
	go srv.Listen(address, port)

	runWebServer()
}

func receiveData(updAddr *net.UDPAddr, data []byte) {
	idx := findSenderIndex(updAddr)
	tools.Logf(3, "Received msg from: %+v, index: %d", *updAddr, idx)
	msg, err := decodeMsg(data)
	if err != nil {
		tools.Logf(0, "decode msg error: %v", err)
		return
	}
	postMsg(&wrapped{
		senderIndex: idx,
		msg:         msg,
	})
}

func findSenderIndex(updAddr *net.UDPAddr) int {
	for i, a := range operators {
		if updAddr.Port == a.Port && updAddr.IP.String() == a.Addr {
			return i
		}
	}
	return -1
}

func inLoop() {
	for msg := range inCh {
		switch msgt := msg.(type) {
		case *wrapped:
			processUDPMsg(msgt)
		case sc.Transaction:
			processTx(msgt)
		}
	}
}

func postMsg(msg interface{}) {
	inCh <- msg
}

func decodeMsg(data []byte) (sc.Transaction, error) {
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

//var suite = bn256.NewSuite()

func processUDPMsg(wrapped *wrapped) {
}

func processTx(tx sc.Transaction) {
	tools.Logf(0, "processTx: id = %s", tx.Id().Short())
	fmt.Printf("%s", tx.ShortStr())

	vtx, _ := tx.ValueTx()
	var buf bytes.Buffer
	err := vtx.Encode().Write(&buf)
	if err != nil {
		return
	}
	sentTo := sendToNodes(buf.Bytes())
	tools.Logf(0, "sent to %+v", sentTo)
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
