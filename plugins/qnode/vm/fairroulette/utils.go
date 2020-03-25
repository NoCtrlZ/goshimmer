package fairroulette

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
)

func getRandom(signature string) uint32 {
	// calculate random uint64
	sigBin, err := hex.DecodeString(signature)
	if err != nil {
		return 0
	}
	h := hashing.HashData(sigBin)
	// rnd < pot, uniformly distributed [0,pot)
	return tools.Uint32From4Bytes(h.Bytes()[:4])
}

func distributePot(ctx vm.RuntimeContext, bets []*BetData, outputs []value.Output) error {
	inputs := make([]*generic.OutputRef, len(bets))
	for i, bet := range bets {
		inputs[i] = bet.OutputRef
	}
	return ctx.SendOutputsToOutputs(inputs, outputs, ctx.SContractAccount())
}

func collectWinners(bets []*BetData, color int) []value.Output {
	if len(bets) == 0 {
		return []value.Output{}
	}
	winningBets := make([]*BetData, 0)
	totalSum := uint64(0)
	winningSum := uint64(0)
	for _, bet := range bets {
		if bet.Color == color {
			winningBets = append(winningBets, bet)
			winningSum += bet.Sum
		}
		totalSum += bet.Sum
	}
	if len(winningBets) == 0 {
		// no bets on the winning color
		return []value.Output{}
	}
	retMap := make(map[hashing.HashValue]value.Output)
	for _, bet := range winningBets {
		if _, ok := retMap[*bet.PayoutAddress]; !ok {
			retMap[*bet.PayoutAddress] = value.NewOutput(bet.PayoutAddress, 0)
		}
		retMap[*bet.PayoutAddress].WithValue(retMap[*bet.PayoutAddress].Value() + bet.Sum)
	}
	// distribute proportionally bets to winning color
	last := value.Output(nil)
	roundedSum := uint64(0)
	for _, outp := range retMap {
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
	ret := make([]value.Output, 0, len(retMap))
	for _, outp := range retMap {
		ret = append(ret, outp)
	}
	return ret
}
