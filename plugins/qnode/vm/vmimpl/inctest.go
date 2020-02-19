package vmimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
)

type inctest struct {
}

func New() vm.Processor {
	return &inctest{}
}

func (_ *inctest) Run(ctx vm.RuntimeContext) {
	reqNr, _ := ctx.InputVars().GetInt("req#")
	stateNr, _ := ctx.OutputVars().GetInt("state#")
	ctx.OutputVars().SetInt("state#", stateNr+1)
	ctx.OutputVars().SetInt("req#", reqNr)
}
