package db

import "github.com/iotaledger/goshimmer/packages/database"

const dbPrefixQnode = byte(100) // TODO proper db prefix code

func Get() (database.Database, error) {
	return database.Get(dbPrefixQnode, database.GetBadgerInstance())
}
