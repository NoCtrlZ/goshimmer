package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
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

type stateResponse struct {
	Account     string         `json:"account"`
	AllBalances []*accountInfo `json:"all_balances"`
	Bets        []*accountInfo `json:"bets"`
}

func newAccountHandler(w http.ResponseWriter, r *http.Request) {
	addr := newAccount()
	resp := &accountInfo{
		Amount:  value.GetBalance(addr),
		Account: addr.String(),
	}
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%v", err))
	}
	_, _ = w.Write(data)
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
	resp := stateResponse{
		AllBalances: getAllBalances(),
		Bets:        getBets(),
	}
	sort.Sort(sortByBalance(resp.AllBalances))
	sort.Sort(sortByBalance(resp.Bets))
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("%v", err))
	}
	_, _ = w.Write(data)
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

func getBets() []*accountInfo {
	tx := getSCState()
	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	if tx == nil {
		return []*accountInfo{}
	}
	betStr, _ := tx.MustState().Vars().GetString("bets")
	if betStr == "" {
		return []*accountInfo{}
	}
	ret := make([]*accountInfo, 0)
	err := json.Unmarshal([]byte(betStr), &ret)
	if err != nil {
		return []*accountInfo{}
	}
	return ret
}

func placeBetHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	sumStr := r.FormValue("sum")
	bet, err := strconv.Atoi(sumStr)
	if err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	bet = bet
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
