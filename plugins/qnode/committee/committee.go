package committee

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"time"
)

// interface to committee

type Committee interface {
	ScId() sctransaction.ScId
	Size() uint16
	OwnPeerIndex() uint16
	SetOperational()
	Dismiss()
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time)
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
}

var New func(scdata *registry.SCData) (Committee, error)
