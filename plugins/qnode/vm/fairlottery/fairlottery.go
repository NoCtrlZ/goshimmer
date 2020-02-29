package fairlottery

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
	"sort"
	"strconv"
	"strings"
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
	vtype, ok := ctx.RequestVars().GetInt("req_type")
	if !ok {
		ctx.SetError(fmt.Errorf("'req_type' undefined"))
		return
	}
	minimumBet, _ := ctx.ConfigVars().GetInt("minimum_bet")
	minimumPot, _ := ctx.ConfigVars().GetInt("minimum_pot")
	if minimumPot < 1000000 {
		// 1 MIOTA
		minimumPot = 1000000 // minimum default
	}
	bets, ok := ctx.StateVars().GetString("bets")
	pot, _ := ctx.StateVars().GetInt("pot")
	numBets, _ := ctx.StateVars().GetInt("num_bets")
	lockedBets, _ := ctx.StateVars().GetString("locked_bets")
	signature, _ := ctx.StateVars().GetString("lock_signature")

	if lockedBets != "" && signature == "" {
		// lockedBets != "" and signature != "" next state update after
		// the lock up of bets
		signature = hex.EncodeToString(ctx.Signature())
		ctx.StateVars().SetString("signature", signature)
	}

	switch vtype {
	case REQ_TYPE_BET:
		// adds another bet in the end of "bets" string
		payoutAddr, ok := ctx.RequestVars().GetString("payout_addr")
		if !ok {
			ctx.SetError(fmt.Errorf("'payout_addr' undefined"))
			return
		}
		depoOutIdx, depositValue := ctx.GetDepositOutput()
		if int(depositValue) < minimumBet {
			ctx.SetError(fmt.Errorf("bet is too small, taken as a donation"))
			return
		}
		bets += fmt.Sprintf("%s,%d,%d,%s|", ctx.RequestTransferId().String(), depoOutIdx, depositValue, payoutAddr)
		pot += int(depositValue) // TODO not correct with types !!!
		ctx.StateVars().SetInt("num_bets", numBets+1)
		ctx.StateVars().SetInt("pot", pot)
		ctx.StateVars().SetString("bets", bets)

	case REQ_TYPE_LOCK:
		// subsequent locks just add up
		if numBets == 0 {
			ctx.SetError(fmt.Errorf("no bets to lock"))
			return
		}
		ctx.StateVars().SetString("locked_bets", lockedBets+bets)
		ctx.StateVars().SetString("locked_signature", "")
		ctx.StateVars().SetString("bets", "")
		ctx.StateVars().SetInt("locked_pot", pot)
		ctx.StateVars().SetInt("num_bets", 0)

		// TODO add Play request to itself

	case REQ_TYPE_PLAY:
		if lockedBets == "" {
			ctx.SetError(fmt.Errorf("no locked bets to play"))
			return
		}
		winner, pot, err := runLottery(lockedBets, signature)
		if err != nil {
			ctx.SetError(err)
			return
		}
		// TODO transfer pot -> winner
		ctx.StateVars().SetString("locked_bets", "")
		ctx.StateVars().SetString("winning_address", winner.String())
		ctx.StateVars().SetInt("payout", int(pot))
	}
}

type betData struct {
	outRef     *generic.OutputRef
	value      uint64
	payoutAddr *hashing.HashValue
}

func scanBetData(lockedBets string) (map[hashing.HashValue]uint64, uint64, error) {
	splittedBets := strings.Split(lockedBets, "|")
	if len(splittedBets) == 0 {
		return nil, 0, fmt.Errorf("no locked bets found")
	}
	bets := make([]*betData, 0, len(splittedBets))
	for _, betStr := range splittedBets {
		betParts := strings.Split(betStr, ",")
		if len(betParts) != 4 {
			return nil, 0, fmt.Errorf("internal inconsistency I")
		}
		transferIdStr := betParts[0]
		depoOutIdxStr := betParts[1]
		depoValueStr := betParts[2]
		payoutAddrStr := betParts[3]

		transferId, err := hashing.HashValueFromString(transferIdStr)
		if err != nil {
			return nil, 0, err
		}
		depoOutIdx, err := strconv.Atoi(depoOutIdxStr)
		if err != nil {
			return nil, 0, err
		}
		payoutAddr, err := hashing.HashValueFromString(payoutAddrStr)
		if err != nil {
			return nil, 0, err
		}
		value, err := strconv.ParseInt(depoValueStr, 10, 64)
		if err != nil {
			return nil, 0, err
		}
		bets = append(bets, &betData{
			outRef:     generic.NewOutputRef(transferId, uint16(depoOutIdx)),
			value:      uint64(value),
			payoutAddr: payoutAddr,
		})
	}
	ret := make(map[hashing.HashValue]uint64)
	pot := uint64(0)
	for _, bd := range bets {
		v, _ := ret[*bd.payoutAddr]
		ret[*bd.payoutAddr] = v + bd.value
		pot += bd.value
	}
	return ret, pot, nil
}

type arrToSort []*betData

func (s arrToSort) Len() int {
	return len(s)
}

func (s arrToSort) Less(i, j int) bool {
	return bytes.Compare((*s[i].payoutAddr).Bytes(), (*s[j].payoutAddr).Bytes()) < 0
}

func (s arrToSort) Swap(i, j int) {
	s[i].payoutAddr, s[j].payoutAddr = s[j].payoutAddr, s[i].payoutAddr
	s[i].value, s[j].value = s[j].value, s[i].value
}

func sortByPayout(bets map[hashing.HashValue]uint64) []*betData {
	toSort := make([]*betData, 0, len(bets))
	for addr, val := range bets {
		toSort = append(toSort, &betData{
			outRef:     nil,
			value:      val,
			payoutAddr: addr.Clone(),
		})
	}
	sort.Sort(arrToSort(toSort))
	return toSort
}

func runLottery(lockedBets, signature string) (*hashing.HashValue, uint64, error) {
	// assert signature != ""
	byPayout, pot, err := scanBetData(lockedBets)
	if err != nil {
		return nil, 0, err
	}
	sortedByPayout := sortByPayout(byPayout)
	// calculate random uint64
	sigBin, err := hex.DecodeString(signature)
	if err != nil {
		return nil, 0, err
	}
	h := hashing.HashData(sigBin)
	// rnd < pot, uniformly distributed [0,pot)
	rnd := tools.Uint64From8Bytes(h.Bytes()[:8]) % pot
	// run roulette
	var runSum uint64
	for _, bd := range sortedByPayout {
		if runSum <= rnd && rnd <= runSum+bd.value {
			return bd.payoutAddr, pot, nil
		}
		runSum += bd.value
	}
	return nil, 0, fmt.Errorf("runLottery: inconsistency")
}
