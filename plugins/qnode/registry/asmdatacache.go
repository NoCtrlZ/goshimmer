package registry

import (
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/hive.go/database"
	"sync"
)

var (
	assemblyDataCache map[HashValue]*AssemblyData
	assemblyDataMutex = &sync.Mutex{}
)

func RefreshAssemblyData() error {
	assemblyDataMutex.Lock()
	defer assemblyDataMutex.Unlock()

	assemblyDataCache = make(map[HashValue]*AssemblyData)
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	err = dbase.ForEachPrefix(dbOpdataGroupKey(), func(entry database.Entry) bool {
		opdata := &AssemblyData{}
		if err = json.Unmarshal(entry.Value, opdata); err == nil {
			assemblyDataCache[*opdata.AssemblyId] = opdata
		}
		return false
	})
	return err
}

func GetAssemblyData(aid *HashValue) (*AssemblyData, bool) {
	assemblyDataMutex.Lock()
	defer assemblyDataMutex.Unlock()
	ret, ok := assemblyDataCache[*aid]
	if !ok {
		return nil, false
	}
	return ret, true
}
