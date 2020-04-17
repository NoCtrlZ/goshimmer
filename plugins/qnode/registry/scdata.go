package registry

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/hive.go/database"
	"sync"
)

var (
	scDataCache map[HashValue]*SCData
	scDataMutex = &sync.Mutex{}
)

func RefreshScData() error {
	scDataMutex.Lock()
	defer scDataMutex.Unlock()

	scDataCache = make(map[HashValue]*SCData)
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	err = dbase.ForEachPrefix(dbOpdataGroupKey(), func(entry database.Entry) bool {
		opdata := &SCData{}
		if err = json.Unmarshal(entry.Value, opdata); err == nil {
			// skip legacy records with Scid == nil
			if opdata.Scid != nil {
				scDataCache[*opdata.Scid] = opdata
			}
		}
		return false
	})
	return err
}

func GetScData(aid *HashValue) (*SCData, bool) {
	scDataMutex.Lock()
	defer scDataMutex.Unlock()
	ret, ok := scDataCache[*aid]
	if !ok {
		return nil, false
	}
	return ret, true
}

func dbOpdataGroupKey() []byte {
	return []byte("opdata")
}

func dbOpdateKey(aid *HashValue) []byte {
	var buf bytes.Buffer
	buf.Write(dbOpdataGroupKey())
	buf.Write(aid.Bytes())
	return buf.Bytes()
}

func (ad *SCData) Save() error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(ad)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbOpdateKey(ad.Scid),
		Value: jsonData,
	})
}

func (sc *SCId) GetSC() (database.Entry, error) {
	var entry database.Entry
	dbase, err := db.Get()
	if err != nil {
		return entry, err
	}
	return dbase.Get(dbOpdateKey(sc.Scid))
}