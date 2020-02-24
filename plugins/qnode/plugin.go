package qnode

import (
	"github.com/iotaledger/goshimmer/packages/parameter"
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/modelimpl"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/signedblock"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

func init() {
	modelimpl.InitModelImplementation()
	signedblock.InitSignedBlockImplementation()
}

const name = "Qnode"

var PLUGIN = node.NewPlugin(name, node.Enabled, config)

func config(_ *node.Plugin) {
	initLoggers()
	qserver.StartServer()
	api.InitEndpoints()
	logParams()
}

func logParams() {
	log := logger.NewLogger(name)
	log.Infow("Qnode plugin parameters:",
		"UDP port",
		parameter.NodeConfig.GetInt(parameters.UDP_PORT),
		"node emulator IP addr",
		parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_IP_ADDR),
		"node emulator IP port",
		parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PORT),
		"web api ",
		parameter.NodeConfig.GetString(webapi.BIND_ADDRESS),
	)
}

func initLoggers() {
	operator.InitLogger()
	dkgapi.InitLogger()
	admapi.InitLogger()
	registry.InitLogger()
}
