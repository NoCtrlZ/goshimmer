package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"net/http"
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

func GetSCdata(addr string, port int, schash *hashing.HashValue) (string, error) {
	url := fmt.Sprintf("http://%s:%d/adm/scdata/%s", addr, port, schash.String())
	fmt.Println(url)
	resp, err := http.Get(url)
	fmt.Println(resp)
	if err != nil {
		return "", err
	}

	var result utils.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	if result.Error != "" {
		err = errors.New(result.Error)
	}

	return "result", err
}
