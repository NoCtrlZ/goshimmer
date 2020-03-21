package qnode

import (
	"github.com/iotaledger/goshimmer/packages/parameter"
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/events"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/modelimpl"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/signedblock"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools/mockclientlib"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const name = "Qnode"

var (
	PLUGIN = node.NewPlugin(name, node.Enabled, config)
	log    *logger.Logger
)

func initModules() {
	modelimpl.Init()
	signedblock.Init()
	messaging.Init()
	events.Init(log)
	mockclientlib.InitMockedValueTangle(log)
}

func config(_ *node.Plugin) {
	log = logger.NewLogger(name)
	initLoggers()
	initModules()

	err := registry.RefreshAssemblyData()
	if err != nil {
		log.Panicf("StartServer::LoadAllAssemblyData %v", err)
		return
	}
	registry.LogLoadedConfigs()

	api.InitEndpoints()
	logParams()
}

func logParams() {
	log.Infow("Qnode plugin parameters:",
		"Messaging port",
		parameter.NodeConfig.GetInt(parameters.QNODE_PORT),
		"Tx pub port",
		parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_PUB_TX_PORT),
		"node emulator IP addr",
		parameter.NodeConfig.GetString(parameters.MOCK_TANGLE_SERVER_IP_ADDR),
		"node emulator IP port",
		parameter.NodeConfig.GetInt(parameters.MOCK_TANGLE_SERVER_PORT),
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
