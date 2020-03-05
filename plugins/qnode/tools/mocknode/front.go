package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
	"net/http"
	"path"
	"sort"
	"strconv"
	"sync"
)

var (
	currentState      sc.Transaction
	currentStateMutex = &sync.Mutex{}
	accounts          = make([]*hashing.HashValue, 0)
)

const (
	Mi          = uint64(1000000)
	Gi          = 1000 * Mi
	Ti          = 1000 * Gi
	depositInit = 61 * Ti
	stdReward   = uint64(2000)
)

func setSCState(tx sc.Transaction) {
	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	currentState = tx
}

func getSCState() sc.Transaction {
	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	return currentState
}

func newAccount() *hashing.HashValue {
	addr := hashing.RandomHash(nil)
	generateAccountWithDeposit(addr, depositInit)

	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	accounts = append(accounts, addr)
	return addr
}

const rootDir = "C:/Users/evaldas/Documents/proj/Go/src/github.com/lunfardo314/goshimmer/plugins/qnode/tools/mocknode/pages"

func staticPageHandler(w http.ResponseWriter, r *http.Request) {
	req := r.URL.Path[len("/static/"):]
	pathname := path.Join(rootDir, req)
	http.ServeFile(w, r, pathname)
}

type accountInfo struct {
	Amount  uint64 `json:"amount"`
	Account string `json:"account"`
}

func placeBetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("placeBetHandler\n")
	var err error
	if err = r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	myAccountStr := r.FormValue("my_account")
	var myAccount *hashing.HashValue

	myAccount, err = hashing.HashValueFromString(myAccountStr)
	if err != nil {
		_, _ = fmt.Fprintf(w, "%v", err)
		return
	}
	myBalance := value.GetBalance(myAccount)
	if myBalance == 0 {
		_, _ = fmt.Fprintf(w, "0 balance")
		return
	}
	sumInt, err := strconv.Atoi(r.FormValue("sum"))
	sum := uint64(sumInt)
	if err != nil || sum == 0 || myBalance < sum+1+stdReward {
		_, _ = fmt.Fprintf(w, "wrong bet amount or not enough balance")
		return
	}
	color, err := strconv.Atoi(r.FormValue("color"))
	if err != nil || color >= fairroulette.NUM_COLORS {
		_, _ = fmt.Fprintf(w, "wrong color code")
		return
	}
	tx, err := makeBetRequestTx(myAccount, sum, color, stdReward)
	if err != nil {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("error: %v", err))
		return
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("error: %v", err))
		return
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("error: %v", err))
		return
	}
	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          tx,
	})
}

type stateResponse struct {
	MyAccount   accountInfo                  `json:"my_account"`
	ScAccount   accountInfo                  `json:"sc_account"`
	NumBets     int                          `json:"num_bets"`
	SumBets     uint64                       `json:"sum_bets"`
	Bets        []*fairroulette.BetData      `json:"bets"`
	ColorStats  [fairroulette.NUM_COLORS]int `json:"color_stats"`
	NumRuns     int                          `json:"num_runs"`
	AllBalances []*accountInfo               `json:"all_balances"`
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("getStateHandler\n")
	var err error
	if err = r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	myAccountStr := r.FormValue("my_account")
	var myAccount *hashing.HashValue

	if myAccountStr != "" {
		if myAccount, err = hashing.HashValueFromString(myAccountStr); err != nil {
			_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
	} else {
		myAccount = newAccount()
	}
	resp := getStateResponse(myAccount)
	sort.Sort(sortByBalance(resp.AllBalances))
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%v", err))
	}
	_, _ = w.Write(data)
}

func getStateResponse(myAccount *hashing.HashValue) *stateResponse {
	ret := &stateResponse{
		MyAccount: accountInfo{
			Amount:  value.GetBalance(myAccount),
			Account: myAccount.String(),
		},
		ScAccount:   getScAccount(),
		Bets:        getBets(),
		ColorStats:  getColorStats(),
		AllBalances: getAllBalances(),
	}
	tx := getSCState()
	if tx == nil {
		return ret
	}
	ret.NumBets = len(ret.Bets)
	ret.SumBets = 0
	for _, bet := range ret.Bets {
		ret.SumBets += bet.Value
	}
	ret.NumRuns, _ = tx.MustState().Vars().GetInt("num_runs")

	return ret
}

func getScAccount() accountInfo {
	ret := accountInfo{}
	tx := getSCState()
	if tx == nil {
		return ret
	}
	ret.Account = tx.MustState().Config().AssemblyAccount().String()
	ret.Amount = value.GetBalance(tx.MustState().Config().AssemblyAccount())
	return ret
}

func getAllBalances() []*accountInfo {
	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	ret := make([]*accountInfo, len(accounts))
	for i, addr := range accounts {
		ret[i] = &accountInfo{
			Amount:  value.GetBalance(addr),
			Account: addr.String(),
		}
	}
	return ret
}

func getColorStats() [fairroulette.NUM_COLORS]int {
	var ret [fairroulette.NUM_COLORS]int
	tx := getSCState()
	if tx == nil {
		return ret
	}
	for i := 0; i < fairroulette.NUM_COLORS; i++ {
		n := fmt.Sprintf("color_%d", i)
		ret[i], _ = tx.MustState().Vars().GetInt(generic.VarName(n))
	}
	return ret
}

func getBets() []*fairroulette.BetData {
	tx := getSCState()
	if tx == nil {
		return []*fairroulette.BetData{}
	}
	betStr, _ := tx.MustState().Vars().GetString("bets")
	if betStr == "" {
		return []*fairroulette.BetData{}
	}
	ret := make([]*fairroulette.BetData, 0)
	err := json.Unmarshal([]byte(betStr), &ret)
	if err != nil {
		return []*fairroulette.BetData{}
	}
	return ret
}

type sortByBalance []*accountInfo

func (s sortByBalance) Len() int {
	return len(s)
}

func (s sortByBalance) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

func (s sortByBalance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
