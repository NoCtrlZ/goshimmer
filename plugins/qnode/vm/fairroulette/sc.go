package fairroulette

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
)

type fairRoulette struct {
}

func New() vm.Processor {
	return &fairRoulette{}
}

const (
	REQ_TYPE_BET        = 1
	REQ_TYPE_LOCK       = 2
	REQ_TYPE_DISTRIBUTE = 3
	NUM_COLORS          = 7
	MAX_BETS            = 20 // when number of bets reaches the limit, lock msg is sent automatically.
)

type BetData struct {
	*generic.OutputRef
	Sum           uint64             `json:"s"`
	Color         int                `json:"c"`
	PayoutAddress *hashing.HashValue `json:"p"`
}

func (_ *fairRoulette) Run(ctx vm.RuntimeContext) {
	// request type is taken from the request variable 'req_type'
	reqType, ok := ctx.RequestVars().GetInt("req_type")
	if !ok {
		ctx.SetError(fmt.Errorf("'req_type' undefined"))
		return
	}
	ctx.Log().Debugf("Req: %d StateVars: %+v", reqType, ctx.StateVars())

	// taking values from the config section
	//minimumBet, _ := ctx.ConfigVars().GetInt("minimum_bet")

	// current yet unlocked bets are stored as string in variable 'bets'
	// as hex encoded json marshaled list of deposit outputs
	// of BET requests
	betsStr, _ := ctx.StateVars().GetString("bets")

	bets := make([]*BetData, 0)
	if betsStr != "" {
		if err := json.Unmarshal([]byte(betsStr), &bets); err != nil {
			ctx.SetError(err)
			return
		}
	}

	// locked bets are stored as string in variable 'locked_bets'
	// as hex encoded json marshaled list of deposit outputs
	lockedBetsStr, _ := ctx.StateVars().GetString("locked_bets")
	lockedBets := make([]*BetData, 0)
	if lockedBetsStr != "" {
		if err := json.Unmarshal([]byte(lockedBetsStr), &lockedBets); err != nil {
			ctx.SetError(err)
			return
		}
	}
	winningColor, ok := ctx.StateVars().GetInt("winning_color")
	if !ok {
		winningColor = -1
	}
	// 'signature' is signature saved next request right after lock request
	// if not locked and right after lock request it is == ""
	signature, _ := ctx.StateVars().GetString("locked_signature")

	if len(lockedBets) != 0 && signature == "" {
		// lockedBets != "" and signature != "" next state update after the lock up of bets
		// this way 'signature' contains signature of state update, produced by the LOCK request
		signature = hex.EncodeToString(ctx.Signature())
		ctx.StateVars().SetString("locked_signature", signature)
		rnd32 := getRandom(signature)
		ctx.StateVars().SetString("rnd", fmt.Sprintf("%d", rnd32))
		winningColor = int(rnd32 % NUM_COLORS)
		ctx.StateVars().SetInt("winning_color", winningColor)
	}
	ctx.StateVars().SetInt("req_type", reqType)

	switch reqType {
	case REQ_TYPE_BET:
		// bet request
		// adds bet to the list unlocked yet bets
		color, ok := ctx.RequestVars().GetInt("color")
		if !ok || color < 0 || color >= NUM_COLORS {
			ctx.SetError(fmt.Errorf("wrong color code"))
			return
		}
		depositOutput := ctx.MainRequestOutputs().DepositOutput
		if depositOutput == nil {
			ctx.SetError(fmt.Errorf("deposit not found"))
			return
		}
		//if int(depositOutput.Value) < minimumBet {
		//	ctx.SetError(fmt.Errorf("bet is too small, taken as donation"))
		//	return
		//}
		payoutAddr, err := ctx.MainInputAddress()
		if err != nil {
			ctx.Log().Errorf("MainInputAddress returned: %v. Assuming payout to the smart contract account", err)
			payoutAddr = ctx.SContractAccount()
		}
		bets = append(bets, &BetData{
			OutputRef:     &depositOutput.OutputRef,
			Sum:           depositOutput.Value,
			Color:         color,
			PayoutAddress: payoutAddr,
		})
		betsBin, err := json.Marshal(bets)
		if err != nil {
			ctx.SetError(err)
			return
		}
		ctx.StateVars().SetString("bets", string(betsBin))

		if len(bets) == MAX_BETS {
			err = ctx.AddRequestToSelf(REQ_TYPE_LOCK)
			if err != nil {
				ctx.SetError(err)
				return
			}
		}

	case REQ_TYPE_LOCK:
		// LOCK request
		// appends list of not locked yet bets to the end of 'locked_bets' list
		// repeats previous state if len(bets) == 0
		if len(bets) != 0 {
			lockedBets = append(lockedBets, bets...)
			lockedBetsBin, err := json.Marshal(lockedBets)
			if err != nil {
				ctx.SetError(err)
				return
			}
			ctx.StateVars().SetString("locked_bets", string(lockedBetsBin))
			ctx.StateVars().SetString("locked_signature", "")
			ctx.StateVars().SetString("bets", "")
			ctx.StateVars().SetString("rnd", "")
			ctx.StateVars().SetInt("winning_color", -1)

			err = ctx.AddRequestToSelf(REQ_TYPE_DISTRIBUTE)
			if err != nil {
				ctx.SetError(err)
				return
			}
		}

	case REQ_TYPE_DISTRIBUTE:
		if len(lockedBets) == 0 || winningColor < 0 || winningColor >= NUM_COLORS {
			ctx.SetError(fmt.Errorf("can't distribute anything"))
			return
		}

		outputs := collectWinners(lockedBets, winningColor)
		if len(outputs) > 0 {
			err := distributePot(ctx, lockedBets, outputs)
			if err != nil {
				ctx.SetError(err)
				return
			}
		}
		// len(outputs) == 0 means nobody staked on the winning color
		// the total stays in the smart contract's account

		ctx.StateVars().SetString("locked_bets", "")

		keyName := generic.VarName(fmt.Sprintf("color_%d", winningColor))
		colorCounter, _ := ctx.StateVars().GetInt(keyName)
		ctx.StateVars().SetInt(keyName, colorCounter+1)

		numRuns, _ := ctx.StateVars().GetInt("num_runs")
		ctx.StateVars().SetInt("num_runs", numRuns+1)

	default:
		ctx.SetError(fmt.Errorf("wrong request type %d", reqType))
		return
	}
}
