package messaging

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"net"
)

type Messaging interface {
	GetOwnAddressAndPort() (string, int)
	SendUDPData([]byte, *hashing.HashValue, uint16, byte, *net.UDPAddr) error
	PostToValueTangle(value.Transaction) error
}
