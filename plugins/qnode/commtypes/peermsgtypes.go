package commtypes

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/peering"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"time"
)

const (
	MsgNotifyRequests         = 0 + peering.FirstCommitteeMsgCode
	MsgStartProcessingRequest = 1 + peering.FirstCommitteeMsgCode
	MsgSignedHash             = 2 + peering.FirstCommitteeMsgCode
	MsgGetStateUpdate         = 3 + peering.FirstCommitteeMsgCode
	MsgStateUpdate            = 4 + peering.FirstCommitteeMsgCode
)

type TimerTick int

// message is sent to the leader of the state processing
// it is sent upon state change or upon arrival of the new request
// the receiving operator will ignore repeating messages
type NotifyReqMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	StateIndex uint32
	// list of request ids ordered by the time of arrival
	RequestIds []sctransaction.RequestId
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request and sign the result hash with the timestamp proposed by the leader
type StartProcessingReqMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// state index in the context of which the message is sent
	StateIndex uint32
	// request id
	RequestId sctransaction.RequestId
}

// after calculations the result peer responds to the start processing msg
// with SignedHashMsg, which contains result hash and signatures
type SignedHashMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	StateIndex uint32
	// timestamp of this message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// request id
	RequestId sctransaction.RequestId
	// original timestamp, the parameter for calculations, which is signed as part of the essence
	OrigTimestamp time.Time
	// hash of the signed data (essence)
	DataHash hashing.HashValue
	// signatures
	//SigBlocks []generic.SignedBlock
}

// request state update from peer. Used in syn process
type GetStateUpdateMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index of the requested state update
	StateIndex uint32
}

// state update sent to peer. Used in sync process
type StateUpdateMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state update
	StateUpdate state.StateUpdate
	// locally calculated by VM (needed for syncing)
	FromVM bool
}

// state manager notifies consensus operator about changed state
// only sent internally within committee
// state transition is always from state N to state N+1
type StateTransitionMsg struct {
	// new variable state
	VariableState state.VariableState
	// corresponding state transaction
	StateTransaction *sctransaction.Transaction
}
