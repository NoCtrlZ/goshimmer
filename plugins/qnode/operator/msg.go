package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/pkg/errors"
	"io"
)

// peer messages between operators
type pushResultMsg struct {
	SenderIndex    uint16
	RequestId      *HashValue
	MasterDataHash *HashValue
	StateIndex     uint32
	SigBlocks      []generic.SignedBlock
}

type pullResultMsg struct {
	SenderIndex uint16
	RequestId   *HashValue
	StateIndex  uint32
	HaveVotes   uint16
}

type timerMsg int

const (
	msgTypePush = messaging.FIRST_SC_MSG_TYPE
	msgTypePull = msgTypePush + 1
)

func (op *scOperator) receiveMsgData(senderIndex uint16, msgType byte, msgData []byte) {
	switch msgType {
	case msgTypePush:
		msg, err := decodePushResultMsg(msgData)
		if err != nil {
			log.Errorf("receiveMsgData: error while decoding push mseesage: %v", err)
			return
		}
		msg.SenderIndex = senderIndex
		op.postEventToQueue(msg)

	case msgTypePull:
		msg, err := decodePullResultMsg(msgData)
		if err != nil {
			log.Errorf("receiveMsgData: error while decoding pull mseesage: %v", err)
			return
		}
		msg.SenderIndex = senderIndex
		op.postEventToQueue(msg)

	default:
		log.Errorf("receiveMsgData: wrong msg type")
	}
}

func encodePushResultMsg(msg *pushResultMsg, buf *bytes.Buffer) {
	buf.Write(msg.RequestId.Bytes())
	buf.Write(msg.MasterDataHash.Bytes())
	_ = tools.WriteUint32(buf, msg.StateIndex)
	_ = generic.WriteSignedBlocks(buf, msg.SigBlocks)
}

func decodePushResultMsg(data []byte) (*pushResultMsg, error) {
	rdr := bytes.NewReader(data)
	var reqId HashValue
	_, err := rdr.Read(reqId.Bytes())
	if err != nil {
		return nil, err
	}
	var masterDataHash HashValue
	_, err = rdr.Read(masterDataHash.Bytes())
	if err != nil {
		return nil, err
	}
	var stateIndex uint32
	err = tools.ReadUint32(rdr, &stateIndex)
	if err != nil {
		return nil, err
	}
	sigBlocks, err := generic.ReadSignedBlocks(rdr)
	if err != nil {
		return nil, err
	}
	return &pushResultMsg{
		MasterDataHash: &masterDataHash,
		RequestId:      &reqId,
		StateIndex:     stateIndex,
		SigBlocks:      sigBlocks,
	}, nil
}

func encodePullResultMsg(msg *pullResultMsg, w io.Writer) {
	_, _ = w.Write(msg.RequestId.Bytes())
	_ = tools.WriteUint32(w, msg.StateIndex)
	_ = tools.WriteUint16(w, msg.HaveVotes)
}

var unexp2 = errors.New("decodePullResultMsg: unexpected end of buffer")

func decodePullResultMsg(data []byte) (*pullResultMsg, error) {
	rdr := bytes.NewReader(data)
	var reqId HashValue
	_, err := rdr.Read(reqId.Bytes())
	if err != nil {
		return nil, unexp2
	}

	var stateIndex uint32
	err = tools.ReadUint32(rdr, &stateIndex)
	if err != nil {
		return nil, unexp2
	}

	var haveVotes uint16
	err = tools.ReadUint16(rdr, &haveVotes)
	if err != nil {
		return nil, unexp2
	}

	ret := &pullResultMsg{
		RequestId:  &reqId,
		StateIndex: stateIndex,
		HaveVotes:  haveVotes,
	}
	return ret, nil
}
