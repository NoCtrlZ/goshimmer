package value

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

// transaction received from the ValueTangle ontology

type Transaction interface {
	Id() *hashing.HashValue
	Transfer() UTXOTransfer
	Payload() []byte
	Encode() generic.Encode
}

type UTXOTransfer interface {
	Id() *hashing.HashValue
	Inputs() []Input
	Outputs() []Output
	AddInput(Input) uint16
	AddOutput(Output) uint16
	InputSignatures() ([]generic.SignedBlock, error)
	DataHash() *hashing.HashValue
	Encode() generic.Encode
	ShortStr() string
}

type Input interface {
	OutputRef() *generic.OutputRef
	GetInputTransfer() UTXOTransfer
	Encode() generic.Encode
}

type Output interface {
	Address() *hashing.HashValue
	Value() uint64
	WithValue(uint64) Output
	Encode() generic.Encode
}
