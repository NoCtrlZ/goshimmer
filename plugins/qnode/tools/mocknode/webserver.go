package main

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"net/http"
)

func runWebServer() {
	fmt.Printf("Web server is running on port %d\n", webport)

	http.HandleFunc("/", startPageHandler)
	http.HandleFunc("/static/", staticPageHandler)
	http.HandleFunc("/demo/game", gamePageHandler)
	http.HandleFunc("/demo/state", getStateHandler)
	http.HandleFunc("/demo/bet", placeBetHandler)

	panic(http.ListenAndServe(fmt.Sprintf(":%d", webport), nil))
}

var originPosted = false

func postOriginIfNeeded() {
	if originPosted {
		return
	}
	tx, err := newOrigin(ownerAddress)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	originPosted = true
	ownerTxPosted = true
	fmt.Printf("origin posted for assembly %s\n", tx.MustState().AssemblyId().Short())

	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          ownerTx,
	})

	postMsg(&wrapped{
		senderIndex: qserver.MockTangleIdx,
		tx:          tx,
	})
}
