package registry

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/registry"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}

func LogLoadedConfigs() {
	assemblyDataMutex.Lock()
	defer assemblyDataMutex.Unlock()

	log.Debugf("loaded %d assembly data record(s)", len(assemblyDataCache))
	for aid, ad := range assemblyDataCache {
		log.Debugw("assembly record", "aid", aid.Short(), "dscr", ad.Description)
	}
}
