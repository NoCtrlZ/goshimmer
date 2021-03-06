package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/pkg/errors"
	"io"
	"net"
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
	MSG_PUSH_MSG = byte(1)
	MSG_PULL_MSG = byte(2)
)

func encodePushResultMsg(msg *pushResultMsg, buf *bytes.Buffer) {
	buf.Write(msg.RequestId.Bytes())
	buf.Write(msg.MasterDataHash.Bytes())
	_ = tools.WriteUint32(buf, msg.StateIndex)
	_ = generic.WriteSignedBlocks(buf, msg.SigBlocks)
}

func decodePushResultMsg(senderIndex uint16, data []byte) (*pushResultMsg, error) {
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
		SenderIndex:    senderIndex,
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

func decodePullResultMsg(senderIndex uint16, data []byte) (*pullResultMsg, error) {
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
		SenderIndex: senderIndex,
		RequestId:   &reqId,
		StateIndex:  stateIndex,
		HaveVotes:   haveVotes,
	}
	return ret, nil
}

func (op *AssemblyOperator) encodeMsg(msg interface{}) ([]byte, byte) {
	var encodedMsg bytes.Buffer
	var typ byte
	switch msgt := msg.(type) {
	case *pushResultMsg:
		encodePushResultMsg(msgt, &encodedMsg)
		typ = MSG_PUSH_MSG
	case *pullResultMsg:
		encodePullResultMsg(msgt, &encodedMsg)
		typ = MSG_PULL_MSG
	default:
		panic("wrong message type")
	}
	return encodedMsg.Bytes(), typ
}

func (op *AssemblyOperator) sendMsgToPeer(msg interface{}, index int16) error {
	if index < 0 || int(index) >= len(op.peers) {
		return errors.New("sendMsgToPeer: wrong peer index")
	}
	encodedMsg, typ := op.encodeMsg(msg)
	return op.comm.SendUDPData(encodedMsg, op.assemblyId, op.peerIndex(), typ, op.peers[index])
}

func (op *AssemblyOperator) sendMsgToPeers(msg interface{}) {
	encodedMsg, typ := op.encodeMsg(msg)
	for _, a := range op.peers {
		if a != nil {
			if err := op.comm.SendUDPData(encodedMsg, op.assemblyId, op.peerIndex(), typ, a); err != nil {
				log.Errorw("SendUDPData", "addr", a.IP.String(), "port", a.Port, "err", err)
			}
		}
	}
}

func (op *AssemblyOperator) validSender(udpAddr *net.UDPAddr, senderIndex uint16) bool {
	if senderIndex < 0 || senderIndex >= op.assemblySize() || senderIndex == op.peerIndex() {
		return false
	}
	return op.peers[senderIndex].IP.String() == udpAddr.IP.String() && op.peers[senderIndex].Port == udpAddr.Port
}
