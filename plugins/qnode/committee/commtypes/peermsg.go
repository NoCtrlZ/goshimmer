package commtypes

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/peering"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
	"time"
)

type timerMsg int

const (
	MsgNotifyRequests         = 0 + peering.FirstCommitteeMsgCode
	MsgStartProcessingRequest = 1 + peering.FirstCommitteeMsgCode
	MsgSignedHash             = 2 + peering.FirstCommitteeMsgCode
	MsgGetStateUpdate         = 3 + peering.FirstCommitteeMsgCode
	MsgStateUpdate            = 4 + peering.FirstCommitteeMsgCode
)

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

func (msg *NotifyReqMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(msg.RequestIds))); err != nil {
		return err
	}
	for _, reqid := range msg.RequestIds {
		if _, err := w.Write(reqid.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (msg *NotifyReqMsg) Read(r io.Reader) error {
	err := util.ReadUint32(r, &msg.StateIndex)
	if err != nil {
		return err
	}
	var arrLen uint16
	err = util.ReadUint16(r, &arrLen)
	if err != nil {
		return err
	}
	if arrLen == 0 {
		return nil
	}
	msg.RequestIds = make([]sctransaction.RequestId, arrLen)
	for i := range msg.RequestIds {
		_, err = r.Read(msg.RequestIds[i].Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (msg *StartProcessingReqMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	_, err := w.Write(msg.RequestId.Bytes())
	return err
}

func (msg *StartProcessingReqMsg) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &msg.StateIndex); err != nil {
		return err
	}
	_, err := r.Read(msg.RequestId.Bytes())
	return err
}

func (msg *SignedHashMsg) Write(w io.Writer) error {
	if err := util.WriteUint32(w, msg.StateIndex); err != nil {
		return err
	}
	if err := util.WriteTime(w, msg.OrigTimestamp); err != nil {
		return err
	}
	if _, err := w.Write(msg.RequestId.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(msg.DataHash.Bytes()); err != nil {
		return err
	}
	//_ = generic.WriteSignedBlocks(buf, msg.SigBlocks)
	return nil
}

func (msg *SignedHashMsg) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &msg.StateIndex); err != nil {
		return err
	}
	if err := util.ReadTime(r, &msg.OrigTimestamp); err != nil {
		return err
	}
	if _, err := r.Read(msg.RequestId.Bytes()); err != nil {
		return err
	}
	if _, err := r.Read(msg.DataHash.Bytes()); err != nil {
		return err
	}
	//ret.SigBlocks, err = generic.ReadSignedBlocks(rdr)
	//if err != nil {
	//	return nil, err
	//}
	return nil
}

func (msg *GetStateUpdateMsg) Write(w io.Writer) error {
	return util.WriteUint32(w, msg.StateIndex)
}

func (msg *GetStateUpdateMsg) Read(r io.Reader) error {
	return util.ReadUint32(r, &msg.StateIndex)
}

func (msg *StateUpdateMsg) Write(w io.Writer) error {
	if err := msg.StateUpdate.Write(w); err != nil {
		return err
	}
	return util.WriteBoolByte(w, msg.FromVM)
}

func (msg *StateUpdateMsg) Read(r io.Reader) error {
	msg.StateUpdate = state.NewStateUpdate(sctransaction.NilScId, 0)
	if err := msg.StateUpdate.Read(r); err != nil {
		return err
	}
	return util.ReadBoolByte(r, &msg.FromVM)
}
