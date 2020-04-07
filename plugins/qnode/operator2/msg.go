package operator2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

type timerMsg int

const (
	// send when state changes or when new request arrives
	// to notify the leader about requests this node has
	msgNotifyRequests = messaging.FirstCommitteeMsgType
	msgProcessRequest = msgNotifyRequests + 1
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
type processReqMsg struct {
	// is set upon receive the message
	SenderIndex uint16
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// state index in the context of which the message is sent
	StateIndex uint32
	// request id
	RequestId *sc.RequestId
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

func encodeProcessReqMsg(msg *processReqMsg, buf *bytes.Buffer) {
	_ = tools.WriteUint32(buf, msg.StateIndex)
	buf.Write(msg.RequestId.Bytes())
}

func decodeProcessReqMsg(data []byte) (*processReqMsg, error) {
	ret := &processReqMsg{}
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

	case msgProcessRequest:
		msg, err := decodeProcessReqMsg(msgData)
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
