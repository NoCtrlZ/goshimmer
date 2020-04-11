package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

type timerMsg int

const (
	msgNotifyRequests         = messaging.FirstCommitteeMsgCode
	msgStartProcessingRequest = msgNotifyRequests + 1
	msgSignedHash             = msgStartProcessingRequest + 1
)

// message is sent to the leader of the state processing
// it is sent upon state change or upon arrival of the new request
// the receiving operator will ignore repeating messages
type notifyReqMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	StateIndex uint32
	// list of request ids ordered by the time of arrival
	RequestIds []*sc.RequestId
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request and sign the result hash with the timestamp proposed by the leader
type startProcessingReqMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// state index in the context of which the message is sent
	StateIndex uint32
	// request id
	RequestId *sc.RequestId
}

// after calculations the result peer responds to the start processing msg
// with signedHashMsg, which contains result hash and signatures
type signedHashMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	StateIndex uint32
	// timestamp of this message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// request id
	RequestId *sc.RequestId
	// original timestamp, the parameter for calculations, which is signed as part of the essence
	OrigTimestamp time.Time
	// hash of the signed data (essence)
	DataHash *hashing.HashValue
	// signatures
	SigBlocks []generic.SignedBlock
}

func encodeNotifyReqMsg(msg *notifyReqMsg, buf *bytes.Buffer) {
	_ = tools.WriteUint32(buf, msg.StateIndex)
	_ = tools.WriteUint16(buf, uint16(len(msg.RequestIds)))
	for _, reqid := range msg.RequestIds {
		buf.Write(reqid.Bytes())
	}
}

func decodeNotifyReqMsg(data []byte) (*notifyReqMsg, error) {
	ret := &notifyReqMsg{}
	rdr := bytes.NewReader(data)
	err := tools.ReadUint32(rdr, &ret.StateIndex)
	if err != nil {
		return nil, err
	}
	var arrLen uint16
	err = tools.ReadUint16(rdr, &arrLen)
	if err != nil {
		return nil, err
	}
	if arrLen == 0 {
		return ret, nil
	}
	ret.RequestIds = make([]*sc.RequestId, arrLen)
	for i := range ret.RequestIds {
		ret.RequestIds[i] = new(sc.RequestId) // can't believe I'm using 'new' :))
		_, err = rdr.Read(ret.RequestIds[i].Bytes())
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func encodeProcessReqMsg(msg *startProcessingReqMsg, buf *bytes.Buffer) {
	_ = tools.WriteUint32(buf, msg.StateIndex)
	buf.Write(msg.RequestId.Bytes())
}

func decodeProcessReqMsg(data []byte) (*startProcessingReqMsg, error) {
	ret := &startProcessingReqMsg{}
	rdr := bytes.NewReader(data)
	err := tools.ReadUint32(rdr, &ret.StateIndex)
	if err != nil {
		return nil, err
	}
	ret.RequestId = new(sc.RequestId)
	_, err = rdr.Read(ret.RequestId.Bytes())
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func encodeSignedHashMsg(msg *signedHashMsg, buf *bytes.Buffer) {
	_ = tools.WriteUint32(buf, msg.StateIndex)
	_ = tools.WriteTime(buf, msg.OrigTimestamp)
	buf.Write(msg.RequestId.Bytes())
	buf.Write(msg.DataHash.Bytes())
	_ = generic.WriteSignedBlocks(buf, msg.SigBlocks)
}

func decodeSignedHashMsg(data []byte) (*signedHashMsg, error) {
	ret := &signedHashMsg{}
	rdr := bytes.NewReader(data)
	err := tools.ReadUint32(rdr, &ret.StateIndex)
	if err != nil {
		return nil, err
	}
	err = tools.ReadTime(rdr, &ret.OrigTimestamp)
	if err != nil {
		return nil, err
	}
	ret.RequestId = new(sc.RequestId)
	_, err = rdr.Read(ret.RequestId.Bytes())
	if err != nil {
		return nil, err
	}
	ret.DataHash = new(hashing.HashValue)
	_, err = rdr.Read(ret.DataHash.Bytes())
	if err != nil {
		return nil, err
	}
	ret.SigBlocks, err = generic.ReadSignedBlocks(rdr)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (op *scOperator) receiveMsgData(senderIndex uint16, msgType byte, msgData []byte, ts time.Time) {
	switch msgType {
	case msgNotifyRequests:
		msg, err := decodeNotifyReqMsg(msgData)
		if err != nil {
			log.Error(err)
			return
		}
		msg.SenderIndex = senderIndex
		op.postEventToQueue(msg)

	case msgStartProcessingRequest:
		msg, err := decodeProcessReqMsg(msgData)
		if err != nil {
			log.Error(err)
			return
		}
		msg.SenderIndex = senderIndex
		msg.Timestamp = ts
		op.postEventToQueue(msg)

	case msgSignedHash:
		msg, err := decodeSignedHashMsg(msgData)
		if err != nil {
			log.Error(err)
			return
		}
		msg.SenderIndex = senderIndex
		msg.Timestamp = ts
		op.postEventToQueue(msg)

	default:
		log.Errorf("receiveMsgData: wrong msg type")
	}
}
