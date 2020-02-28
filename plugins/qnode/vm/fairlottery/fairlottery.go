package fairlottery

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
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
	REQ_TYPE_BET     = 1
	REQ_TYPE_LOCK    = 2
	REQ_TYPE_REWARDS = 3
)

func (_ *fairLottery) Run(ctx vm.RuntimeContext) {
	vars := ctx.InputVars()
	outVars := ctx.OutputVars()
	vtype, ok := vars.GetInt("req_type")
	if !ok {
		ctx.SetError(fmt.Errorf("'req_type' undefined"))
		return
	}

	switch vtype {
	case REQ_TYPE_BET:
		payoutAddr, ok := vars.GetString("payout_addr")
		if !ok {
			ctx.SetError(fmt.Errorf("'payout_addr' undefined"))
			return
		}
		bets, ok := vars.GetString("bets")
		if !ok {
			bets = ""
		}
		depoOutIdx, depoValue := ctx.GetDepositOutput()
		bets += fmt.Sprintf("%s,%d,%d,%s|", ctx.RequestTransferId().String(), depoOutIdx, depoValue, payoutAddr)
		outVars.SetString("bets", bets)

	case REQ_TYPE_LOCK:
		lockedBets, ok := vars.GetString("locked_bets")
		if !ok {
			lockedBets = ""
		}
		bets, ok := vars.GetString("bets")
		if !ok {
			bets = ""
		}
		lockedBets += bets
		outVars.SetString("bets", "")
		outVars.SetString("locked_bets", lockedBets)

	case REQ_TYPE_REWARDS:
		lockedBets, ok := vars.GetString("locked_bets")
		if !ok || lockedBets == "" {
			ctx.SetError(fmt.Errorf("no locked bets were found"))
			return
		}
		betData, err := scanBetData(lockedBets)
		if err != nil {
			ctx.SetError(err)
			return
		}
		byPayout := sumByPayoutAddr(betData)
		sortedByPayout := sortByPayout(byPayout)
		winner, pot := runLottery(sortedByPayout, ctx.GetRandom())
		// TODO transfer
		outVars.SetString("locked_bets", "")
	}
}

type betData struct {
	outRef     *generic.OutputRef
	value      uint64
	payoutAddr *hashing.HashValue
}

func scanBetData(lockedBets string) ([]*betData, error) {
	splittedBets := strings.Split(lockedBets, "|")
	if len(splittedBets) == 0 {
		return nil, fmt.Errorf("no locked bets found")
	}
	ret := make([]*betData, 0, len(splittedBets))
	for _, betStr := range splittedBets {
		betParts := strings.Split(betStr, ",")
		if len(betParts) != 4 {
			return nil, fmt.Errorf("internal inconsistency I")
		}
		transferIdStr := betParts[0]
		depoOutIdxStr := betParts[1]
		depoValueStr := betParts[2]
		payoutAddrStr := betParts[3]

		transferId, err := hashing.HashValueFromString(transferIdStr)
		if err != nil {
			return nil, err
		}
		depoOutIdx, err := strconv.Atoi(depoOutIdxStr)
		if err != nil {
			return nil, err
		}
		payoutAddr, err := hashing.HashValueFromString(payoutAddrStr)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseInt(depoValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &betData{
			outRef:     generic.NewOutputRef(transferId, uint16(depoOutIdx)),
			value:      uint64(value),
			payoutAddr: payoutAddr,
		})
	}
	return ret, nil
}

func sumByPayoutAddr(betData []*betData) map[hashing.HashValue]uint64 {
	ret := make(map[hashing.HashValue]uint64)
	for _, bd := range betData {
		v, ok := ret[*bd.payoutAddr]
		if !ok {
			ret[*bd.payoutAddr] = bd.value
		} else {
			ret[*bd.payoutAddr] = v + bd.value
		}
	}
	return ret
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

func runLottery(sortedBets []*betData, rnd uint32) (*hashing.HashValue, uint64) {
	// run roulette
	sumPot := uint64(0)
	for _, bd := range sortedBets {
		sumPot += bd.value
	}
	rndAdjusted := uint64(rnd) % sumPot
	if rndAdjusted == 0 {
		return sortedBets[0].payoutAddr, sumPot
	}
	// rndAdjusted < sumPot
	var runSum uint64
	for i, bd := range sortedBets {
		if runSum <= rndAdjusted && rndAdjusted < runSum+bd.value {
			return bd.payoutAddr, sumPot
		}
	}
	return sortedBets[len(sortedBets)-1].payoutAddr, sumPot
}
