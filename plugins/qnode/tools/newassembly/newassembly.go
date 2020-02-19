package main

import (
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/glb"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"io/ioutil"
	"time"
)

const (
	dir = "C://Users//evaldas//Documents//proj//site_data//goshimmer/"
)

func main() {
	od := operator.AssemblyData{
		OwnerPubKey: "ownerPubKey",
		Description: "Test assembly 2",
		Program:     "A+B",
		Modified:    time.Now().UnixNano(),
	}
	var err error
	od.AssemblyId = hash.HashStrings(od.OwnerPubKey, od.Description, od.Program)
	data, err := json.MarshalIndent(&od, " ", "")
	if err != nil {
		panic(err)
	}
	fname := dir + od.AssemblyId.String() + ".json"
	err = ioutil.WriteFile(fname, data, 0644)
	if err != nil {
		panic(err)
	}
}
