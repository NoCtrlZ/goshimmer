package db

import "github.com/iotaledger/goshimmer/packages/database"

const dbPrefixQnode = byte(100)

func Get() (database.Database, error) {
	return database.Get(dbPrefixQnode)
}
