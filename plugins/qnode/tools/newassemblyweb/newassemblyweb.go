package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/glb"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"io/ioutil"
	"net/http"
)

const (
	dir = "C://Users//evaldas//Documents//proj//site_data//goshimmer/"
	aid = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
)

var apiPorts = []int{8080, 8081, 8082, 8083}

func main() {
	dat, err := ioutil.ReadFile(dir + aid + ".json")
	if err != nil {
		panic(err)
	}
	var od operator.AssemblyData
	var aid hash.HashValue
	od.AssemblyId = &aid
	err = json.Unmarshal(dat, &od)
	if err != nil {
		panic(err)
	}

	for _, port := range apiPorts {
		url := fmt.Sprintf("http://localhost:%d/adm/newassembly", port)
		resp, err := http.Post(url, "application/json",
			bytes.NewBuffer(dat))
		if err != nil {
			fmt.Printf("Error occured 1: %v\n", err)
			continue
		}

		var result api.SimpleResponse

		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			fmt.Printf("Error occured 2: %v\n", err)
			continue
		}
		if result.Error == "" {
			fmt.Printf("newassembly request to port %d success\n", port)
		} else {
			fmt.Printf("newassembly request to port %d error %s\n", port, result.Error)
		}
	}
}
