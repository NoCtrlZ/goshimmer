package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
	"math/rand"
	"net/http"
	"strconv"
)

const webport = 2000

func runWebServer() {
	fmt.Printf("Web server is running on port %d\n", webport)
	//http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/testbuttons", testButtonsHandler)
	http.HandleFunc("/testreq", testreqHandler)
	http.HandleFunc("/testbet", betTestHandler)
	http.HandleFunc("/testlock", lockHandler)

	http.HandleFunc("/static/", staticPageHandler)
	http.HandleFunc("/demo/start", startPageHandler)
	http.HandleFunc("/demo/game", gamePageHandler)
	http.HandleFunc("/demo/state", getStateHandler)
	http.HandleFunc("/demo/bet", placeBetHandler)

	panic(http.ListenAndServe(fmt.Sprintf(":%d", webport), nil))
}

func testButtonsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "C:/Users/evaldas/Documents/proj/Go/src/github.com/lunfardo314/goshimmer/plugins/qnode/tools/mocknode/sendmsg.html")
}

var originPosted = false

func testreqHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	var tx sc.Transaction
	var err error
	if !originPosted {
		originPosted = true
		postOrigin()
		return
	}

	reqnr := r.FormValue("reqnr")
	num, err := strconv.Atoi(reqnr)
	if err != nil || num == 0 {
		num = 1
	}
	fmt.Printf("send %d requests\n", num)
	for i := 0; i < num; i++ {
		tx, err = makeReqTx()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		vtx, err := tx.ValueTx()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		if err := ldb.PutTransaction(vtx); err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		postMsg(&wrapped{
			senderIndex: qserver.MockTangleIdx,
			tx:          tx,
		})
	}

}

func postOrigin() {
	tx, err := newOrigin()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	originPosted = true
	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          tx,
	})
}

func betTestHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	betStr := r.FormValue("sum")
	if betStr != "" {
		bet, err := strconv.Atoi(betStr)
		if err != nil {
			fmt.Printf("wrong bet amount '%s'\n", betStr)
			return
		}
		fmt.Printf("Bet request received for %d iotas\n", bet)

		playerIdx := curPlayer
		curPlayer = (curPlayer + 1) % numPlayers
		fromAccount := requesterAddresses[playerIdx]

		tx, err := makeBetRequestTx(fromAccount, uint64(bet), rand.Intn(fairroulette.NUM_COLORS), 2000)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		vtx, err := tx.ValueTx()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		if err := ldb.PutTransaction(vtx); err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		postMsg(&wrapped{
			senderIndex: qserver.MockTangleIdx,
			tx:          tx,
		})
	}
}

func lockHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	tx, err := makeLockRequestTx()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	if err := ldb.PutTransaction(vtx); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          tx,
	})
}
