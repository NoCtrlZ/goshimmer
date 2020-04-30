package commtypes

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"time"
)

// interface for decoupling

type Committee interface {
	ScId() sctransaction.ScId
	Size() uint16
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time)
	IsAlivePeer(peerIndex uint16) bool
	ProcessMessage(msg interface{})
}
