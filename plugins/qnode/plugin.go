package qnode

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/qserver"
	"github.com/iotaledger/hive.go/node"
)

const name = "Qnode"

var PLUGIN = node.NewPlugin(name, node.Enabled, config)

func config(_ *node.Plugin) {
	operator.InitLogger()
	dkgapi.InitLogger()

	qserver.StartServer()
	api.InitEndpoints()
}
