package messaging

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"net"
)

type Messaging interface {
	GetOwnAddressAndPort() (string, int)
	SendUDPData([]byte, *hashing.HashValue, uint16, byte, *net.UDPAddr) error
}
