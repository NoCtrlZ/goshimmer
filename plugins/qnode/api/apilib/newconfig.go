package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"net/http"
)

func NewConfiguration(addr string, port int, cdata *operator.ConfigData) error {
	data, err := json.Marshal(cdata)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%d/adm/newconfig", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	var result utils.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	err = nil
	if result.Error != "" {
		err = errors.New(result.Error)
	}
	return err
}
