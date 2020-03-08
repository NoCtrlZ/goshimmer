package qserver

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/network/udp"
	"github.com/pkg/errors"
	"net"
)

func createUDPServer() *udp.UDPServer {
	ret := udp.NewServer(parameters.UDP_BUFFER_SIZE)

	ret.Events.Start.Attach(events.NewClosure(func() {
		log.Debugf("Start event qnode UDP server")
	}))
	ret.Events.Shutdown.Attach(events.NewClosure(func() {
		log.Debugf("Shutdown event qnode UDP server")
	}))
	ret.Events.ReceiveData.Attach(events.NewClosure(receiveUDPData))

	ret.Events.Error.Attach(events.NewClosure(func(err error) {
		log.Debugf("Error event qnode UDP server: %v", err)
	}))
	return ret
}

func receiveUDPData(updAddr *net.UDPAddr, data []byte) {
	err := receiveUDPDataErr(updAddr, data)
	if err != nil {
		log.Errorf("receiveUDPData: %v", err)
	}
}

const MockTangleIdx = uint16(0xFFFF)

func receiveUDPDataErr(updAddr *net.UDPAddr, data []byte) error {
	aid, senderIndex, msgType, msgData, err := UnwrapUDPPacket(data)
	if err != nil {
		return err
	}
	if senderIndex == MockTangleIdx && aid.Equal(NilHash) && ServerInstance.isMockTangleAddr(updAddr) {
		// for testing: processing messages from MockTangle server
		tx, err := value.ParseTransaction(msgData)
		if err != nil {
			return err
		}

		// -- for testing only TODO
		if err := ServerInstance.txdb.PutTransaction(tx); err != nil {
			return err
		}
		// -- for testing only

		ServerInstance.Events.NodeEvent.Trigger(tx)
		return nil
	}
	ambly, ok := ServerInstance.getOperator(aid)
	if !ok {
		return errors.New("no such assembly")
	}
	return ambly.ReceiveUDPData(updAddr, senderIndex, msgType, msgData)
}

func UnwrapUDPPacket(data []byte) (*HashValue, uint16, byte, []byte, error) {
	rdr := bytes.NewBuffer(data)
	var aid HashValue
	_, err := rdr.Read(aid.Bytes())
	if err != nil {
		return nil, 0, 0, nil, err
	}
	var senderIndex uint16
	err = tools.ReadUint16(rdr, &senderIndex)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	msgType, err := tools.ReadByte(rdr)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return &aid, senderIndex, msgType, rdr.Bytes(), nil
}

func WrapUDPPacket(aid *HashValue, senderIndex uint16, msgType byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.Write(aid.Bytes())
	_ = tools.WriteUint16(&buf, senderIndex)
	buf.WriteByte(msgType)
	buf.Write(data)
	return buf.Bytes()
}
