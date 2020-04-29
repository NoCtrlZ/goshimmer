package main

import (
	"github.com/iotaledger/goshimmer/plugins/banner"
	"github.com/iotaledger/goshimmer/plugins/cli"
	"github.com/iotaledger/goshimmer/plugins/config"
	"github.com/iotaledger/goshimmer/plugins/database"
	"github.com/iotaledger/goshimmer/plugins/gracefulshutdown"
	"github.com/iotaledger/goshimmer/plugins/logger"
	"github.com/iotaledger/goshimmer/plugins/qnode"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/node"
	_ "net/http/pprof"
)

var LOCAL_CORE_PLUGINS = node.Plugins(
	banner.Plugin,
	config.Plugin,
	logger.Plugin,
	cli.Plugin,
	//portcheck.PLUGIN,
	database.Plugin,
	//autopeering.PLUGIN,
	//messagelayer.PLUGIN,
	//gossip.PLUGIN,
	gracefulshutdown.Plugin,
	//metrics.PLUGIN,
	//drng.PLUGIN,
)

var LOCAL_WEBAPI_PLUGINS = node.Plugins(
	webapi.Plugin,
	//webauth.PLUGIN,
	//spammer.PLUGIN,
	//data.PLUGIN,
	//drng.PLUGIN,
	//message.PLUGIN,
	//autopeering.PLUGIN,
	//info.Plugin,
)

func main() {
	// go http.ListenAndServe("localhost:6061", nil) // pprof Server for Debbuging Mutexes

	node.Run(
		LOCAL_CORE_PLUGINS,
		//research.PLUGINS,
		//ui.PLUGINS,
		LOCAL_WEBAPI_PLUGINS,
		qnode.PLUGINS,
	)
}
