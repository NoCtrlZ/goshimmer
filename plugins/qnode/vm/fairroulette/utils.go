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
		inputs[i] = &bet.OutputRef
	}
	return ctx.SendOutputsToOutputs(inputs, outputs, ctx.AssemblyAccount())
}

func collectWinners(bets []*BetData, color int) []value.Output {
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
