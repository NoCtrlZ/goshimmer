package commtypes

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

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
