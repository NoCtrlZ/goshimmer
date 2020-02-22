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
	Equal(Transaction) bool
}

type State interface {
	AssemblyId() *HashValue
	ConfigId() *HashValue
	RequestRef() (*HashValue, uint16)
	StateChainOutputIndex() uint16
	StateChainAccount() *HashValue
	RequestChainAccount() *HashValue
	OwnerChainAccount() *HashValue
	ConfigVars() generic.ValueMap
	StateVars() generic.ValueMap
	StateIndex() uint32
	WithStateIndex(uint32) State
	WithConfigVars(generic.ValueMap) State
	WithStateVars(generic.ValueMap) State
	WithSetStateChainOutputIndex(uint16) State
	Encode() generic.Encode
}

type Request interface {
	AssemblyId() *HashValue
	IsConfigUpdateReq() bool
	Vars() generic.ValueMap
	RequestChainOutputIndex() uint16
	Encode() generic.Encode
}

type RequestRef struct {
	tx           Transaction
	requestIndex uint16
}

type StateUpdateMsg struct {
	Tx Transaction
}
