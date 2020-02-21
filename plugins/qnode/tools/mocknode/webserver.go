package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"net/http"
)

const webport = 2000

func runWebServer() {
	fmt.Printf("Web server is running on port %d\n", webport)
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", webport), nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	var tx sc.Transaction
	var err error
	isCfgStr := r.FormValue("cfg")
	if isCfgStr != "" {
		tx, err = newOrigin()
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: %v", err)
			return
		}
	} else {
		reqnr := r.FormValue("reqnr")
		if reqnr == "" {
			_, _ = fmt.Fprintf(w, "reqnr parameter not provided")
			return
		}
		_, _ = fmt.Fprintf(w, "Received reqnr = %s", reqnr)
		tx, err = makeReqTx(reqnr)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: %v", err)
			return
		}
	}
	postMsg(tx)
}
