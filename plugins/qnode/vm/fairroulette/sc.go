package fairroulette

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
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
)

type betData struct {
	generic.OutputRefWithAddrValue
	Color int `json:"color"`
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
	minimumBet, _ := ctx.ConfigVars().GetInt("minimum_bet")

	// current yet unlocked bets are stored as string in variable 'bets'
	// as hex encoded json marshaled list of deposit outputs
	// of BET requests
	betsStr, _ := ctx.StateVars().GetString("bets")

	bets := make([]*betData, 0)
	if betsStr != "" {
		if err := json.Unmarshal([]byte(betsStr), &bets); err != nil {
			ctx.SetError(err)
			return
		}
	}

	// locked bets are stored as string in variable 'locked_bets'
	// as hex encoded json marshaled list of deposit outputs
	lockedBetsStr, _ := ctx.StateVars().GetString("locked_bets")
	lockedBets := make([]*betData, 0)
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
	// 'num_bets' is a counter of not locked yet bets
	numBets, _ := ctx.StateVars().GetInt("num_bets")
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
		if int(depositOutput.Value) < minimumBet {
			ctx.SetError(fmt.Errorf("bet is too small, taken as donation"))
			return
		}
		bets = append(bets, &betData{
			OutputRefWithAddrValue: *depositOutput,
			Color:                  color,
		})
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
		ctx.StateVars().SetString("locked_bets", string(lockedBetsBin))
		ctx.StateVars().SetString("locked_signature", "")
		ctx.StateVars().SetString("bets", "")
		ctx.StateVars().SetInt("num_bets", 0)
		ctx.StateVars().SetString("rnd", "")
		ctx.StateVars().SetInt("winning_color", -1)

		ctx.AddRequestToSelf(REQ_TYPE_DISTRIBUTE)

	case REQ_TYPE_DISTRIBUTE:
		if len(lockedBets) == 0 || winningColor < 0 || winningColor >= NUM_COLORS {
			ctx.SetError(fmt.Errorf("can't distribute anything"))
			return
		}

		outputs := collectWinners(lockedBets, winningColor)
		err := distributePot(ctx, lockedBets, outputs)
		if err != nil {
			ctx.SetError(err)
			return
		}
		ctx.StateVars().SetString("locked_bets", "")

		keyName := generic.VarName(fmt.Sprintf("color_%d", winningColor))
		colorCounter, _ := ctx.StateVars().GetInt(keyName)
		ctx.StateVars().SetInt(keyName, colorCounter+1)

	default:
		ctx.SetError(fmt.Errorf("wrong request type %d", reqType))
		return
	}
}

func distributePot(ctx vm.RuntimeContext, bets []*betData, outputs []value.Output) error {
	inputs := make([]*generic.OutputRef, len(bets))
	for i, bet := range bets {
		inputs[i] = &bet.OutputRef
	}
	return ctx.SendOutputsToOutputs(inputs, outputs, ctx.AssemblyAccount())
}

func collectWinners(bets []*betData, color int) []value.Output {
	ret1 := make(map[hashing.HashValue]value.Output)
	if len(bets) == 0 {
		return []value.Output{}
	}
	winningBets := make([]*generic.OutputRefWithAddrValue, 0)
	totalSum := uint64(0)
	winningSum := uint64(0)
	for _, bet := range bets {
		if bet.Color == color {
			winningBets = append(winningBets, &bet.OutputRefWithAddrValue)
			winningSum += bet.OutputRefWithAddrValue.Value
		}
		totalSum += bet.OutputRefWithAddrValue.Value
	}
	for _, bet := range winningBets {
		if _, ok := ret1[*bet.Addr]; !ok {
			ret1[*bet.Addr] = value.NewOutput(bet.Addr, 0)
		}
		ret1[*bet.Addr].WithValue(ret1[*bet.Addr].Value() + bet.Value)
	}
	// distribute proportionally bets to winning color
	last := value.NewOutput(hashing.NilHash, 0)
	roundedSum := uint64(0)
	for _, outp := range ret1 {
		last = outp
		coeff := float64(outp.Value()) / float64(winningSum)
		sum := uint64(coeff * float64(totalSum))
		outp.WithValue(sum)
		roundedSum += sum
	}
	// adjust to rounding
	switch {
	case roundedSum > totalSum:
		last.WithValue(last.Value() - (roundedSum - totalSum))
	case roundedSum < totalSum:
		last.WithValue(last.Value() + (totalSum - roundedSum))
	}
	ret := make([]value.Output, 0, len(ret1))
	for _, outp := range ret1 {
		ret = append(ret, outp)
	}
	return ret
}
