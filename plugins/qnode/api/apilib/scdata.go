package apilib

import (
	"fmt"
	"bytes"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
)

func PutSCData(addr string, port int, adata *registry.SCData) error {
	data, err := json.Marshal(adata)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%d/adm/scdata", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	var result utils.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Error != "" {
		err = errors.New(result.Error)
	}
	return err
}

func GetSCdata(addr string, port int, schash *registry.SCId) (string, error) {
	data, err := json.Marshal(schash)
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("http://%s:%d/adm/getsc", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	fmt.Println("request done")
	if err != nil {
		fmt.Println("response error")
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("status code error")
		panic("response is invalid")
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("json parse error")
		panic(err)
	}
	bodyString := string(bodyBytes)
	return bodyString, err
}
