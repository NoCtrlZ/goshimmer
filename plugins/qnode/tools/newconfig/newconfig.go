package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/glb"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/pkg/errors"
	"net/http"
)

var apiPorts = []int{8080, 8081, 8082, 8083}
var udpPorts = []int{4000, 4001, 4002, 4003}

const (
	ipAddress  = "127.0.0.1"
	aidString  = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	dksName    = "distributed key set 3"
	configName = "configuration 4"
	N          = 4
	T          = 3
)

func main() {
	if N != len(apiPorts) {
		panic("wrong params")
	}
	aid, err := hash.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	dksId := hash.HashStrings(dksName)
	configId := hash.HashStrings(configName)

	fmt.Printf("creating new configuration %+v\n", apiPorts)
	fmt.Printf("assembly id = %s\ndistributed set name = '%s'\nid = %s\n", aid, dksName, dksId.String())
	fmt.Printf("assembly id = %s\nconfig name = '%s'\nconfig id = %s\n", aid, configName, configId.String())
	fmt.Println("---------------------------------------")

	req := admapi.NewConfigRequest{
		AssemblyId:    aid,
		DKShareId:     dksId,
		Id:            configId,
		OperatorAddrs: make([]*operator.PortAddr, len(apiPorts)),
	}
	for i, port := range udpPorts {
		req.OperatorAddrs[i] = &operator.PortAddr{
			Port: port,
			Addr: ipAddress,
		}
	}
	for i, port := range apiPorts {
		_, err := newConfig(&req, port)
		if err == nil {
			fmt.Printf("new config (%d, %d) OK\n", i, port)
		} else {
			fmt.Printf("new config (%d, %d) FAIL: %v\n", i, port, err)
		}
	}
}

func newConfig(req *admapi.NewConfigRequest, port int) (*admapi.NewConfigResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/adm/newconfig", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret admapi.NewConfigResponse
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}
