package sctransaction

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

// state block of the SC transaction. Represents SC state update
// previous state block can be determined by the chain transfer of the SC token in the UTXO part of the
// transaction
type StateBlock struct {
	// scid of the SC which is updated
	// scid contains balance.NEW_COLOR in the scid.Color field for the origin transaction
	scid *ScId
	// stata index is 0 for the origin transaction
	// consensus maintains incremental sequence of state indexes
	stateIndex uint32
	// timestamp of the transaction. 0 means transaction is not timestamped
	timestamp int64
	// requestId = tx hash + requestId index which originated this state update
	// the list is needed for batches of requests
	// this reference makes requestIds (inputs to state update) immutable part of the state update
	requestIds []RequestId
	// variable state hash.
	// it is used to validate the variable state in the SC ledger while syncing
	// note that by having this has it is still impossible to reach respective state update without
	// syncing the whole chain
	variableStateHash hashing.HashValue
}

type StateBlockParams struct {
	Timestamp         int64
	RequestIds        []RequestId
	VariableStateHash hashing.HashValue
}

func NewStateBlock(scid *ScId, stateIndex uint32) *StateBlock {
	return &StateBlock{
		scid:       scid,
		stateIndex: stateIndex,
		requestIds: make([]RequestId, 0),
	}
}

// getters/setters

func (sb *StateBlock) ScId() *ScId {
	return sb.scid
}

func (sb *StateBlock) StateIndex() uint32 {
	return sb.stateIndex
}

func (sb *StateBlock) Timestamp() int64 {
	return sb.timestamp
}

func (sb *StateBlock) RequestIds() *[]RequestId {
	return &sb.requestIds
}

func (sb *StateBlock) VariableStateHash() hashing.HashValue {
	return sb.variableStateHash
}

func (sb *StateBlock) WithParams(params StateBlockParams) *StateBlock {
	sb.timestamp = params.Timestamp
	sb.requestIds = params.RequestIds
	sb.variableStateHash = params.VariableStateHash
	return sb
}

// encoding
// important: each block starts with 65 bytes of scid

func (sb *StateBlock) Write(w io.Writer) error {
	if err := sb.scid.Write(w); err != nil {
		return err
	}
	if err := util.WriteUint32(w, sb.stateIndex); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(sb.timestamp)); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(sb.requestIds))); err != nil {
		return err
	}
	for i := range sb.requestIds {
		if err := sb.requestIds[i].Write(w); err != nil {
			return err
		}
	}
	if err := sb.variableStateHash.Write(w); err != nil {
		return err
	}
	return nil
}

func (sb *StateBlock) Read(r io.Reader) error {
	scid := new(ScId)
	if err := scid.Read(r); err != nil {
		return err
	}
	var stateIndex uint32
	if err := util.ReadUint32(r, &stateIndex); err != nil {
		return err
	}
	var timestamp uint64
	if err := util.ReadUint64(r, &timestamp); err != nil {
		return err
	}
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	requestIds := make([]RequestId, size)
	for i := range sb.requestIds {
		if err := sb.requestIds[i].Read(r); err != nil {
			return err
		}
	}
	var stateUpdateHash hashing.HashValue
	if err := stateUpdateHash.Read(r); err != nil {
		return err
	}
	sb.scid = scid
	sb.stateIndex = stateIndex
	sb.timestamp = int64(timestamp)
	sb.requestIds = requestIds
	sb.variableStateHash = stateUpdateHash
	return nil
}
