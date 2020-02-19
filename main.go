package main

import (
	"github.com/iotaledger/goshimmer/plugins/qnode"
	"net/http"
	_ "net/http/pprof"

	"github.com/iotaledger/goshimmer/plugins/cli"
	"github.com/iotaledger/goshimmer/plugins/gracefulshutdown"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/node"
)

func main() {
	cli.LoadConfig()

	go http.ListenAndServe("localhost:6060", nil) // pprof Server for Debbuging Mutexes

	node.Run(
		node.Plugins(
			cli.PLUGIN,
			//remotelog.PLUGIN,
			//
			//autopeering.PLUGIN,
			//gossip.PLUGIN,
			//tangle.PLUGIN,
			//bundleprocessor.PLUGIN,
			//analysis.PLUGIN,
			gracefulshutdown.PLUGIN,
			//tipselection.PLUGIN,
			//metrics.PLUGIN,

			webapi.PLUGIN,
			qnode.PLUGIN,
			//webapi_auth.PLUGIN,
			//webapi_gtta.PLUGIN,
			//webapi_spammer.PLUGIN,
			//webapi_broadcastData.PLUGIN,
			//webapi_getTransactionTrytesByHash.PLUGIN,
			//webapi_getTransactionObjectsByHash.PLUGIN,
			//webapi_findTransactionHashes.PLUGIN,
			//webapi_getNeighbors.PLUGIN,
			//webapi_spammer.PLUGIN,

			//spa.PLUGIN,
			//graph.PLUGIN,
		),
	)
}
