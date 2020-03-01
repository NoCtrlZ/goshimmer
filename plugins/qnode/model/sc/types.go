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
	Error() error
	Vars() generic.ValueMap
	StateChainOutputIndex() uint16
	WithError(error) State
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
	MAP_KEY_ASSEMBLY_ACCOUNT  = "assembly_account"
	MAP_KEY_OWNER_ACCOUNT     = "owner_account"
	MAP_KEY_CHAIN_OUT_INDEX   = "chain_out_idx"
	MAP_KEY_REWARD_OUT_INDEX  = "reward_out_idx"
	MAP_KEY_DEPOSIT_OUT_INDEX = "chain_out_idx"
)

type Request interface {
	AssemblyId() *HashValue
	IsConfigUpdateReq() bool
	Vars() generic.ValueMap
	MainOutputs(Transaction) [3]*generic.OutputRefWithAddrValue
	WithRequestChainOutputIndex(uint16) Request
	WithRewardOutputIndex(uint16) Request
	WithDepositOutputIndex(uint16) Request
	Encode() generic.Encode
}

type RequestRef struct {
	tx           Transaction
	requestIndex uint16
}

type StateUpdateMsg struct {
	Tx Transaction
}
