package messaging

import (
	"bytes"
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"time"
)

// structure of the message:
// timestamp   8 bytes
// msg type    1 byte
// is type != 0, next:
// scid 32 bytes
// sender index 2 bytes
// data variable bytes

type unwrappedPacket struct {
	ts          int64
	msgType     byte
	scid        *HashValue
	senderIndex uint16
	data        []byte
}

func unmarshalPacket(data []byte) (*unwrappedPacket, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("too short message")
	}
	rdr := bytes.NewBuffer(data)
	var uts uint64
	err := tools.ReadUint64(rdr, &uts)
	if err != nil {
		return nil, err
	}
	ret := &unwrappedPacket{
		ts: int64(uts),
	}
	ret.msgType, err = tools.ReadByte(rdr)
	if err != nil {
		return nil, err
	}
	if ret.msgType == 0 {
		return ret, nil
	}
	// committee message
	var scid HashValue
	_, err = rdr.Read(scid.Bytes())
	if err != nil {
		return nil, err
	}
	ret.scid = &scid
	err = tools.ReadUint16(rdr, &ret.senderIndex)
	if err != nil {
		return nil, err
	}
	ret.data = rdr.Bytes()
	return ret, nil
}

// also puts timestamp

func marshalPacket(up *unwrappedPacket) ([]byte, time.Time) {
	var buf bytes.Buffer
	// puts timestamp
	ts := time.Now()
	_ = tools.WriteUint64(&buf, uint64(ts.UnixNano()))
	if up == nil || up.msgType == 0 {
		buf.WriteByte(0)
		return buf.Bytes(), ts
	}
	buf.WriteByte(up.msgType)
	buf.Write(up.scid.Bytes())
	_ = tools.WriteUint16(&buf, up.senderIndex)
	buf.Write(up.data)
	return buf.Bytes(), ts
}
