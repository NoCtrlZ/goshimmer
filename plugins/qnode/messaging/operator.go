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
	ReceiveStateUpdate(*sc.StateUpdateMsg)
	ReceiveRequest(*sc.RequestRef)
	IsDismissed() bool
}
