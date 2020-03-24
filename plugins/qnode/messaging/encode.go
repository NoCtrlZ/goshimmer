package messaging

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func unwrapPacket(data []byte) (*HashValue, uint16, byte, []byte, error) {
	rdr := bytes.NewBuffer(data)
	var aid HashValue
	_, err := rdr.Read(aid.Bytes())
	if err != nil {
		return nil, 0, 0, nil, err
	}
	var senderIndex uint16
	err = tools.ReadUint16(rdr, &senderIndex)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	msgType, err := tools.ReadByte(rdr)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return &aid, senderIndex, msgType, rdr.Bytes(), nil
}

func wrapPacket(aid *HashValue, senderIndex uint16, msgType byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.Write(aid.Bytes())
	_ = tools.WriteUint16(&buf, senderIndex)
	buf.WriteByte(msgType)
	buf.Write(data)
	return buf.Bytes()
}
