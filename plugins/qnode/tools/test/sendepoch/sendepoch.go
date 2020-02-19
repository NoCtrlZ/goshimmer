package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/testapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"net/http"
)

var apiPorts = []int{8080, 8081, 8082, 8083}

const (
	aidString  = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	configName = "configuration 4"
	epochName  = "epoch 2"
)

func main() {
	aid, err := hashing.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	configId := hashing.HashStrings(configName)
	epochId := hashing.HashStrings(epochName)

	fmt.Printf("sending mock epoch to %+v\n", apiPorts)
	fmt.Printf("assembly id = %s\nconfig name = '%s'\nid = %s\n", aid, configName, configId.String())
	fmt.Printf("epoch name = '%s'\nepoch id = %s\n", epochName, epochId.String())
	fmt.Println("---------------------------------------")
	req := &testapi.MockEpochRequest{
		AssemblyId:             aid,
		ConfigId:               configId,
		Id:                     epochId,
		Enabled:                true,
		RequestNotificationLen: 3,
	}
	for i, port := range apiPorts {
		_, err := sendEpoch(req, port)
		if err == nil {
			fmt.Printf("mock epoch (%d, %d) OK\n", i, port)
		} else {
			fmt.Printf("mock epoch (%d, %d) FAIL: %v\n", i, port, err)
		}
	}
}

func sendEpoch(req *testapi.MockEpochRequest, port int) (*testapi.MockEpochResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/testapi/mockepoch", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret testapi.MockEpochResponse
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil

}
