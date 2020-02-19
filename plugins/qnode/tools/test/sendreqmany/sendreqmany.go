package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/testapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strconv"
	"time"
)

var apiPorts = []int{8080, 8082, 8081, 8083}

const (
	aidString = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	dataStr   = "2+2"
)

var numRequests = 100

func main() {
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err != nil || n <= 0 {
			fmt.Printf("wrong parameter")
			os.Exit(1)
		} else {
			numRequests = n
		}
	}
	fmt.Printf("sending %d random mock requests to %+v\n", numRequests, apiPorts)
	fmt.Printf("assembly id = %s\ndata string = '%s'\n", aidString, dataStr)
	fmt.Println("---------------------------------------")
	for i := 0; i < numRequests; i++ {
		sendRndRequest(i)
	}
}

func sendRndRequest(idx int) {
	aid, err := hashing.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	req := &testapi.MockRequestMsgReq{
		AssemblyId: aid,
		Timestamp:  time.Now().UnixNano(),
		Parameters: map[string]interface{}{"#": idx},
		//Data:       hex.EncodeToString([]byte(dataStr)),
	}
	req.Id = hashing.RandomHash(nil)

	for i, port := range apiPorts {
		_, err := sendRequest(req, port)
		if err == nil {
			fmt.Printf("#%d mock request to (%d, %d) OK: id = %s\n", idx, i, port, req.Id.Short())
		} else {
			fmt.Printf("#%d mock request (%d, %d) FAIL: %v\n", idx, i, port, err)
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
