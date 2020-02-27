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
	Signatures() []generic.SignedBlock
	MasterDataHash() *HashValue
	ShortStr() string
	SetState(State)
	AddRequest(Request)
	Equal(Transaction) bool
}

type State interface {
	AssemblyId() *HashValue
	ConfigId() *HashValue
	RequestId() *HashValue
	StateChainOutputIndex() uint16 // TODO maybe just select first or any with 1i output to the StateChainAccount
	StateChainAccount() *HashValue
	RequestChainAccount() *HashValue
	OwnerChainAccount() *HashValue
	ConfigVars() generic.ValueMap
	StateVars() generic.ValueMap
	StateIndex() uint32
	WithStateIndex(uint32) State
	WithConfigVars(generic.ValueMap) State
	WithStateVars(generic.ValueMap) State
	WithStateChainOutputIndex(uint16) State
	Encode() generic.Encode
}

const (
	MAP_KEY_STATE_ACCOUNT   = "state_addr"
	MAP_KEY_REQUEST_ACCOUNT = "request_addr"
	MAP_KEY_OWNER_ACCOUNT   = "owner_addr"
)

type Request interface {
	AssemblyId() *HashValue
	IsConfigUpdateReq() bool
	Vars() generic.ValueMap
	RequestChainOutputIndex() uint16 // TODO maybe just select first or any with 1i output to the RequestChainAccount
	Encode() generic.Encode
}

type RequestRef struct {
	tx           Transaction
	requestIndex uint16
}

type StateUpdateMsg struct {
	Tx Transaction
}
