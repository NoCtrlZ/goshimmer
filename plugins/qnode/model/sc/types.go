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
	RequestId() *HashValue
	StateIndex() uint32
	Vars() generic.ValueMap
	StateChainOutputIndex() uint16
	WithStateIndex(uint32) State
	WithVars(generic.ValueMap) State
	WithStateChainOutputIndex(uint16) State
	Config() Config
	Encode() generic.Encode
}

type Config interface {
	Id() *HashValue
	Vars() generic.ValueMap
	AssemblyAccount() *HashValue
	OwnerAccount() *HashValue
	MinimumReward() uint64
	OwnersMargin() byte // owner's take in percents
	With(Config) Config
}

const (
	MAP_KEY_ASSEMBLY_ACCOUNT = "assembly_account"
	MAP_KEY_OWNER_ACCOUNT    = "owner_account"
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
