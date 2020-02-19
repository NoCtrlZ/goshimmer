package sc

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

type Transaction interface {
	Transfer() value.UTXOTransfer
	ValueTx() (value.Transaction, error)
	Id() *HashValue
	State() (State, bool)
	MustState() State
	Requests() []Request
	Signatures() map[HashValue]generic.SignedBlock
	MasterDataHash() *HashValue
	ShortStr() string
	SetState(State)
	AddRequest(Request)
}

type State interface {
	AssemblyId() *HashValue
	ConfigId() *HashValue
	RequestRef() (*HashValue, uint16)
	StateChainOutputIndex() uint16
	StateChainAddress() *HashValue
	RequestChainAddress() *HashValue
	OwnerChainAddress() *HashValue
	ConfigVars() generic.ValueMap
	StateVars() generic.ValueMap
	StateIndex() uint32
	Builder() StateBuilder
	Encode() generic.Encode
}

type StateBuilder interface {
	InitFromPrev(State)
	SetConfigVars(generic.ValueMap)
	SetStateVars(generic.ValueMap)
	SetRequest(*HashValue, uint16)
	SetStateChainOutputIndex(uint16)
}

type Request interface {
	AssemblyId() *HashValue
	IsConfigUpdateReq() bool
	Vars() generic.ValueMap
	RequestChainOutputIndex() uint16
	Encode() generic.Encode
}

type RequestMsg struct {
	Tx           Transaction
	RequestIndex uint16
}

type StateUpdateMsg struct {
	Tx Transaction
}
