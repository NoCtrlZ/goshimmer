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
	"html/template"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
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
	depositInit = 1 * Gi
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

func getAccount(seed string) *hashing.HashValue {
	h := hashing.HashStrings(seed)
	currentStateMutex.Lock()
	defer currentStateMutex.Unlock()
	for _, addr := range accounts {
		if h.Equal(addr) {
			return h
		}
	}
	generateAccountWithDeposit(h, depositInit)
	accounts = append(accounts, h)
	return h
}

const rootDir = "C:/Users/evaldas/Documents/proj/Go/src/github.com/lunfardo314/goshimmer/plugins/qnode/tools/mocknode/pages"

var mainTemplate *template.Template

func startPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(rootDir, "front.html"))
}

func gamePageHandler(w http.ResponseWriter, r *http.Request) {
	if mainTemplate == nil {
		mainTemplate = template.Must(template.ParseFiles(path.Join(rootDir, "game.html")))
	}
	var err error
	if err = r.ParseForm(); err != nil {
		respondErr(w, err.Error())
		return
	}

	mySeed := strings.TrimSpace(r.FormValue("seed"))
	if len(mySeed) < 5 {
		respondErr(w, "seed must be at least 5 characters")
		return
	}
	_ = mainTemplate.Execute(w, &struct{ Seed string }{Seed: mySeed})
}

func staticPageHandler(w http.ResponseWriter, r *http.Request) {
	req := r.URL.Path[len("/static/"):]
	pathname := path.Join(rootDir, req)
	http.ServeFile(w, r, pathname)
}

type accountInfo struct {
	Amount  uint64 `json:"amount"`
	Account string `json:"account"`
}

type simpleResponse struct {
	Err string `json:"err"`
}

func respondErr(w io.Writer, err string) {
	data, _ := json.Marshal(&simpleResponse{Err: err})
	_, _ = w.Write(data)
}

func placeBetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("placeBetHandler\n")
	if !originPosted {
		originPosted = true
		postOrigin()
		return
	}

	var err error
	if err = r.ParseForm(); err != nil {
		respondErr(w, err.Error())
		return
	}
	mySeed := r.FormValue("seed")
	if len(mySeed) < 5 {
		respondErr(w, "wrong seed")
		return
	}

	myAccount := getAccount(mySeed)
	myBalance := value.GetBalance(myAccount)
	if myBalance == 0 {
		respondErr(w, "0 balance")
		return
	}
	sumInt, err := strconv.Atoi(r.FormValue("sum"))
	sum := uint64(sumInt)
	if err != nil || sum == 0 || myBalance < sum+1+stdReward {
		respondErr(w, "wrong bet amount or not enough balance")
		return
	}
	color, err := strconv.Atoi(r.FormValue("color"))
	if err != nil || color >= fairroulette.NUM_COLORS {
		respondErr(w, "wrong color code")
		return
	}
	tx, err := makeBetRequestTx(myAccount, sum, color, stdReward)
	if err != nil {
		respondErr(w, err.Error())
		return
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		respondErr(w, err.Error())
		return
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		respondErr(w, err.Error())
		return
	}
	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          tx,
	})
	respondErr(w, "")
}

type stateResponse struct {
	MySeed       string                       `json:"my_seed"`
	MyAccount    accountInfo                  `json:"my_account"`
	ScAccount    accountInfo                  `json:"sc_account"`
	NumBets      int                          `json:"num_bets"`
	SumBets      uint64                       `json:"sum_bets"`
	Bets         []*fairroulette.BetData      `json:"bets"`
	WinningColor int                          `json:"winning_color"`
	Sign         string                       `json:"sign"`
	ColorStats   [fairroulette.NUM_COLORS]int `json:"color_stats"`
	NumRuns      int                          `json:"num_runs"`
	AllBalances  []*accountInfo               `json:"all_balances"`
	Err          string                       `json:"err"`
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		respondErr(w, err.Error())
		return
	}

	mySeed := strings.TrimSpace(r.FormValue("seed"))
	if len(mySeed) < 5 {
		respondErr(w, "must be at least 5 characters")
		return
	}
	myAccount := getAccount(mySeed)
	//fmt.Printf("getState for seed '%s' account %s\n", mySeed, myAccount.Short())

	resp := getStateResponse(myAccount)
	resp.MySeed = mySeed
	sort.Sort(sortByBalance(resp.AllBalances))
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		respondErr(w, err.Error())
		return
	}
	_, _ = w.Write(data)
}

func getStateResponse(myAccount *hashing.HashValue) *stateResponse {
	bets, totalStaked := getBets()
	ret := &stateResponse{
		MyAccount: accountInfo{
			Amount:  value.GetBalance(myAccount),
			Account: myAccount.String(),
		},
		ScAccount:   getScAccount(),
		Bets:        bets,
		SumBets:     totalStaked,
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
		ret.SumBets += bet.Sum
	}
	ret.NumRuns, _ = tx.MustState().Vars().GetInt("num_runs")
	ret.WinningColor, _ = tx.MustState().Vars().GetInt("winning_color")
	ret.Sign, _ = tx.MustState().Vars().GetString("locked_signature")
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

func getBets() ([]*fairroulette.BetData, uint64) {
	tx := getSCState()
	if tx == nil {
		return []*fairroulette.BetData{}, 0
	}
	betStr, _ := tx.MustState().Vars().GetString("bets")
	if betStr == "" {
		return []*fairroulette.BetData{}, 0
	}
	ret := make([]*fairroulette.BetData, 0)
	err := json.Unmarshal([]byte(betStr), &ret)
	if err != nil {
		return []*fairroulette.BetData{}, 0
	}
	sum := uint64(0)
	for _, b := range ret {
		sum += b.Sum
	}
	return ret, sum
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
