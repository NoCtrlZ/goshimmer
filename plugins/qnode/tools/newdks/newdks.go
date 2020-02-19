package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"net/http"
)

var apiPorts = []int{8080, 8081, 8082, 8083}

const (
	aidString = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	dksName   = "distributed key set 3"
	N         = 4
	T         = 3
)

func main() {
	if N != len(apiPorts) {
		panic("wrong params")
	}
	aid, err := hashing.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	var dksId = hashing.HashStrings(dksName)

	fmt.Printf("creating new distributed key set at nodes (ports) %+v\n", apiPorts)
	fmt.Printf("assembly id = %s\ndistributed set name = '%s'\nid = %s\n", aid, dksName, dksId.String())
	fmt.Println("---------------------------------------")

	// create new dkshares on respective node instances
	newDKSresponses := make([]*dkgapi.NewDKSResponse, N)
	for index, port := range apiPorts {
		fmt.Printf("new DKS, index %d: calling port %d\n", index, port)
		newDKSresponses[index], err = createDKS(aid, dksId, N, T, index, port)
		if err != nil {
			panic(err)
		}
		fmt.Printf("success new DKS. Index: %d, port %d\n", index, port)
	}

	// aggregates keys
	aggregateReqs := make([]*dkgapi.AggregateDKSRequest, N)
	for i := range aggregateReqs {
		aggregateReqs[i] = &dkgapi.AggregateDKSRequest{
			AssemblyId: aid,
			Id:         dksId,
			Index:      int16(i),
			PriShares:  make([]string, N),
		}
	}
	for i, r := range newDKSresponses {
		for j := range aggregateReqs {
			if j == i {
				aggregateReqs[j].PriShares[r.Index] = ""
			} else {
				aggregateReqs[j].PriShares[r.Index] = r.PriShares[j]
			}
		}
	}
	pubShares := make([]string, N)
	for i := 0; i < N; i++ {
		fmt.Printf("aggregate DKS, index %d: calling port %d\n", i, apiPorts[i])
		a, err := aggregateDKS(aggregateReqs[i], apiPorts[i])
		if err != nil {
			panic(err)
		}
		if a.Err != "" {
			panic(a.Err)
		}
		pubShares[i] = a.PubShare
		fmt.Printf("success aggregate DKS, index %d: calling port %d\n", i, apiPorts[i])
	}
	// commit keys
	for i, port := range apiPorts {
		fmt.Printf("commit DKS, index %d: calling port %d\n", i, port)
		err = commitDKS(aid, dksId, pubShares, port)
		if err != nil {
			panic(err)
		}
		fmt.Printf("success commit DKS, index %d: calling port %d\n", i, port)
	}
	fmt.Println("----------------------------------- success!")
	fmt.Printf("assembly id = %s\n", aid.String())
	fmt.Printf("distributed key set id = %s\n", dksId.String())
}

func createDKS(aid, dksId *hashing.HashValue, n, t, index int, port int) (*dkgapi.NewDKSResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/adm/newdks", port)
	req := dkgapi.NewDKSRequest{
		AssemblyId: aid,
		Id:         dksId,
		N:          n,
		T:          t,
		Index:      int16(index),
	}
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret dkgapi.NewDKSResponse

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}

func aggregateDKS(req *dkgapi.AggregateDKSRequest, port int) (*dkgapi.AggregateDKSResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/adm/aggregatedks", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret dkgapi.AggregateDKSResponse

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}

func commitDKS(aid, dksId *hashing.HashValue, pubShares []string, port int) error {
	req := dkgapi.CommitDKSRequest{
		AssemblyId: aid,
		Id:         dksId,
		PubShares:  pubShares,
	}
	url := fmt.Sprintf("http://localhost:%d/adm/commitdks", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return err
	}

	var ret dkgapi.CommitDKSResponse

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil

}
