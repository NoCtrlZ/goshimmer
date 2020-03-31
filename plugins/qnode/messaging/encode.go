package messaging

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

// structure of the spec message
// type byte + data
// structure of the committee msg
// type byte, 32 bytes scid, 2 bytes sender index, data

func unwrapPacket(data []byte) (*HashValue, uint16, byte, []byte, error) {
	rdr := bytes.NewBuffer(data)
	msgType, err := tools.ReadByte(rdr)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	if msgType < FIRST_SC_MSG_TYPE {
		// special comm message
		return nil, 0, msgType, rdr.Bytes(), nil
	}

	var aid HashValue
	_, err = rdr.Read(aid.Bytes())
	if err != nil {
		return nil, 0, 0, nil, err
	}
	var senderIndex uint16
	err = tools.ReadUint16(rdr, &senderIndex)
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return &aid, senderIndex, msgType, rdr.Bytes(), nil
}

func wrapPacket(aid *HashValue, senderIndex uint16, msgType byte, data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(msgType)
	if msgType >= FIRST_SC_MSG_TYPE {
		// committee message
		buf.Write(aid.Bytes())
		_ = tools.WriteUint16(&buf, senderIndex)
	}
	buf.Write(data)
	return buf.Bytes()
}
