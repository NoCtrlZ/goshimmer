package fairlottery

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
)

type fairLottery struct {
}

func New() vm.Processor {
	return &fairLottery{}
}

const (
	REQ_TYPE_BET  = 1
	REQ_TYPE_LOCK = 2
	REQ_TYPE_PLAY = 3
)

func (_ *fairLottery) Run(ctx vm.RuntimeContext) {
	// request type is taken from the request variable 'req_type'
	vtype, ok := ctx.RequestVars().GetInt("req_type")
	if !ok {
		ctx.SetError(fmt.Errorf("'req_type' undefined"))
		return
	}
	// taking values from the config section
	minimumBet, _ := ctx.ConfigVars().GetInt("minimum_bet")

	// current yet unlocked bets are stored as string in variable 'bets'
	// as hex encoded json marshaled list of deposit outputs
	// of BET requests
	betsStr, _ := ctx.StateVars().GetString("bets")
	ctx.Log().Debugw("", "betStr", betsStr)

	bets := make([]*generic.OutputRefWithAddrValue, 0)
	if betsStr != "" {
		if err := json.Unmarshal([]byte(betsStr), &bets); err != nil {
			ctx.SetError(err)
			return
		}
	}

	// locked bets are stored as string in variable 'locked_bets'
	// as hex encoded json marshaled list of deposit outputs
	lockedBetsStr, _ := ctx.StateVars().GetString("locked_bets")
	lockedBets := make([]*generic.OutputRefWithAddrValue, 0)
	if lockedBetsStr != "" {
		if err := json.Unmarshal([]byte(lockedBetsStr), &lockedBets); err != nil {
			ctx.SetError(err)
			return
		}
	}

	// 'num_bets' is a counter of not locked yet bets
	numBets, _ := ctx.StateVars().GetInt("num_bets")
	// 'signature' is signature saved next request right after lock request
	// if not locked and right after lock request it is == ""
	signature, _ := ctx.StateVars().GetString("lock_signature")
	if len(lockedBets) != 0 && signature == "" {
		// lockedBets != "" and signature != "" next state update after the lock up of bets
		// this way 'signature' contains signature of state update, produced by the LOCK request
		signature = hex.EncodeToString(ctx.Signature())
		ctx.StateVars().SetString("signature", signature)
	}

	switch vtype {
	case REQ_TYPE_BET:
		// bet request
		// adds nre bet to the list unlocked yet bets
		depositOutput := ctx.MainRequestOutputs().DepositOutput
		if depositOutput == nil {
			ctx.SetError(fmt.Errorf("deposit not found"))
			return
		}
		if int(depositOutput.Value) < minimumBet {
			ctx.SetError(fmt.Errorf("bet is too small, taken as donation"))
			return
		}
		bets = append(bets, depositOutput)
		betsBin, err := json.Marshal(bets)
		if err != nil {
			ctx.SetError(err)
			return
		}
		ctx.StateVars().SetString("bets", string(betsBin))
		ctx.StateVars().SetInt("num_bets", numBets+1)

	case REQ_TYPE_LOCK:
		// LOCK request
		// appends list of not locked yet bets to the end of 'locked_bets' list
		if numBets == 0 {
			ctx.SetError(fmt.Errorf("no bets to lock"))
			return
		}
		lockedBets = append(lockedBets, bets...)
		lockedBetsBin, err := json.Marshal(lockedBets)
		if err != nil {
			ctx.SetError(err)
			return
		}
		ctx.StateVars().SetString("locked_bets", hex.EncodeToString(lockedBetsBin))
		ctx.StateVars().SetString("locked_signature", "")
		ctx.StateVars().SetString("bets", "")
		ctx.StateVars().SetInt("num_bets", 0)

		ctx.AddRequestToSelf(REQ_TYPE_PLAY)

	case REQ_TYPE_PLAY:
		// PLAY lottery request
		if len(lockedBets) == 0 {
			ctx.SetError(fmt.Errorf("no locked bets to play"))
			return
		}
		// if there's at least 1 locked bet, "signature" is not empty.
		// It will be used as provably random number
		winner, pot, err := runLottery(lockedBets, signature)
		if err != nil {
			ctx.SetError(err)
			return
		}
		outputs := make([]*generic.OutputRef, len(lockedBets))
		for i, bet := range lockedBets {
			outputs[i] = &bet.OutputRef
		}
		// send all deposit outputs of BET requests to the winning address
		ctx.SendFundsToAddress(outputs, winner)

		ctx.StateVars().SetString("locked_bets", "")
		ctx.StateVars().SetString("winning_address", winner.String())
		ctx.StateVars().SetInt("payout", int(pot))
	default:
		ctx.SetError(fmt.Errorf("wrong request type"))
		return
	}
}
