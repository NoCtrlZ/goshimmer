package qnode

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const name = "Qnode"

var (
	PLUGIN = node.NewPlugin(name, node.Enabled, config)
	log    *logger.Logger
)

func config(_ *node.Plugin) {
	log = logger.NewLogger(name)
	operator.InitLogger()

	err := qserver.StartServer()
	if err != nil {
		log.Errorf("failed to start qnode server: %v", err)
		return
	}
	addApiEndpoints()
}

func addApiEndpoints() {
	webapi.Server.POST("/adm/newdks", dkgapi.HandlerNewDks)
	webapi.Server.POST("/adm/aggregatedks", dkgapi.HandlerAggregateDks)
	webapi.Server.POST("/adm/commitdks", dkgapi.HandlerCommitDks)
	webapi.Server.POST("/adm/signdigest", dkgapi.HandlerSignDigest)
	webapi.Server.POST("/adm/getpubs", dkgapi.HandlerGetPubs)
	webapi.Server.POST("/adm/newconfig", admapi.HandlerNewConfig)
	webapi.Server.POST("/adm/newassembly", admapi.HandlerNewAssembly)

	log.Info("successfully added api endpoints")
}
