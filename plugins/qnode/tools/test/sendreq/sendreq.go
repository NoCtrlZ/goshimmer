package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/testapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var apiPorts = []int{8080, 8081, 8082, 8083}

const (
	aidString = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
)

func main() {
	aid, err := hashing.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	params := make(map[string]interface{})
	params["A"] = 10
	params["B"] = 10
	fmt.Printf("sending mock request to %+v\n", apiPorts)
	fmt.Printf("assembly id = %s\n", aid)
	fmt.Printf("params = %+v\n", params)
	fmt.Println("---------------------------------------")
	req := &testapi.MockRequestMsgReq{
		AssemblyId: aid,
		Id:         hashing.RandomHash(nil),
		Timestamp:  time.Now().UnixNano(),
		Parameters: params,
	}
	for i, port := range apiPorts {
		_, err := sendRequest(req, port)
		if err == nil {
			fmt.Printf("mock request (%d, %d) OK\n", i, port)
		} else {
			fmt.Printf("mock request (%d, %d) FAIL: %v\n", i, port, err)
		}
	}
}

func sendRequest(req *testapi.MockRequestMsgReq, port int) (*testapi.MockRequestMsgResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/testapi/mockreq", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret testapi.MockRequestMsgResponse
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}
