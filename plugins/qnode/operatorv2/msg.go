package operator

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/pkg/errors"
	"io"
)

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
