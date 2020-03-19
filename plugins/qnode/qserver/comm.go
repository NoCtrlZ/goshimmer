package qserver

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"net"
)

// Comm interface

func (q *QServer) GetOwnAddressAndPort() (string, int) {
	return "127.0.0.1", q.udpPort
}

func (q *QServer) SendUDPData(data []byte, aid *hashing.HashValue, senderIndex uint16, msgType byte, addr *net.UDPAddr) error {
	wrapped := WrapUDPPacket(aid, senderIndex, msgType, data)
	if len(wrapped) > parameters.UDP_BUFFER_SIZE {
		return fmt.Errorf("len(wrapped) > parameters.UDP_BUFFER_SIZE. Message wasnt't send")
	}
	_, err := q.udpServer.GetSocket().WriteTo(wrapped, addr)
	return err
}
