package vm

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
)

type Processor interface {
	Run(ctx RuntimeContext)
}

type RuntimeContext interface {
	InputVars() generic.ValueMap
	OutputVars() generic.ValueMap
}
