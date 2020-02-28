package fairlottery

import "github.com/iotaledger/goshimmer/plugins/qnode/vm"

type fairLottery struct {
}

func New() vm.Processor {
	return &fairLottery{}
}

const (
	REQ_TYPE_BET     = 1
	REQ_TYPE_LOCK    = 2
	REQ_TYPE_REWARDS = 3
)

func (_ *fairLottery) Run(ctx vm.RuntimeContext) {
	vars := ctx.InputVars()
	vtype, ok := vars.GetInt("reqType")
	if !ok {
		return
	}
	rewardsLocked, _ := vars.GetBool("rewards_locked")

	switch vtype {
	case REQ_TYPE_BET:
	case REQ_TYPE_LOCK:
	case REQ_TYPE_REWARDS:

	}

	//ctx.OutputVars().SetInt("req#", reqNr)
}
