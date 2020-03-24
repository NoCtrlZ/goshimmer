package messaging

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

type SCOperator interface {
	SContractID() *hashing.HashValue
	Quorum() uint16
	CommitteeSize() uint16
	PeerIndex() uint16
	NodeAddresses() []*registry.PortAddr
	ReceiveMsgData(senderIndex uint16, msgType byte, msgData []byte) error
	ReceiveStateUpdate(msg *sc.StateUpdateMsg)
	ReceiveRequest(msg *sc.RequestRef)
	IsDismissed() bool
}
