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
	reqNr, _ := ctx.RequestVars().GetInt("req#")
	ctx.StateVars().SetInt("req#", reqNr)
}
