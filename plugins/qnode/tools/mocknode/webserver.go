package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"net/http"
	"strconv"
)

const webport = 2000

func runWebServer() {
	fmt.Printf("Web server is running on port %d\n", webport)
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/testreq", testreqHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", webport), nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("defaultHandler\n")
	http.ServeFile(w, r, "C:/Users/evaldas/Documents/proj/Go/src/github.com/lunfardo314/goshimmer/plugins/qnode/tools/mocknode/sendmsg.html")
}

func testreqHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	var tx sc.Transaction
	var err error
	isCfgStr := r.FormValue("cfg")
	if isCfgStr != "" {
		fmt.Printf("cfg request\n")
		tx, err = newOrigin()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		postMsg(&wrapped{
			senderIndex: qserver.MockTangleIdx,
			tx:          tx,
		})
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
		postMsg(&wrapped{
			senderIndex: qserver.MockTangleIdx,
			tx:          tx,
		})
	}

}
