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
	// requestId tx hash + requestId index which originated this state update
	// this reference makes requestId (inputs to state update) immutable part of the state update
	requestId RequestId
	// state hash, referencing the hash of the cached state. It is a tip of the Merkle chain of state updates.
	// it is only used while syncing the state to check weather the state is correct.
	// each moment in time node maintains valid current state of the SC and it can be accessed directly
	stateHash hashing.HashValue
}

type StateBlockParams struct {
	Timestamp       int64
	RequestId       RequestId
	StateUpdateHash hashing.HashValue
}

func NewStateBlock(scid *ScId, stateIndex uint32) *StateBlock {
	return &StateBlock{
		scid:       scid,
		stateIndex: stateIndex,
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

func (sb *StateBlock) RequestId() *RequestId {
	return &sb.requestId
}

func (sb *StateBlock) StateUpdateHash() *hashing.HashValue {
	return &sb.stateHash
}

func (sb *StateBlock) WithParams(params StateBlockParams) *StateBlock {
	sb.timestamp = params.Timestamp
	sb.requestId = params.RequestId
	sb.stateHash = params.StateUpdateHash
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
	if err := sb.requestId.Write(w); err != nil {
		return err
	}
	if err := sb.stateHash.Write(w); err != nil {
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
	var reqId RequestId
	if err := reqId.Read(r); err != nil {
		return err
	}
	var stateUpdateHash hashing.HashValue
	if err := stateUpdateHash.Read(r); err != nil {
		return err
	}
	sb.scid = scid
	sb.stateIndex = stateIndex
	sb.timestamp = int64(timestamp)
	sb.requestId = reqId
	sb.stateHash = stateUpdateHash
	return nil
}
