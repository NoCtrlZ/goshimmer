package admapi

import (
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"fmt"
	"encoding/json"
	"net/http"
)

type SCList []*registry.SCData

type GetScListResponse struct {
	SCList       `json:"contracts"`
	Error string `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	var req GetScListResponse
	scid := hashing.HashData([]byte{1, 2, 3, 4, 5})
	ownerpub := hashing.HashData([]byte{6, 7, 8, 9, 10})
	dscr := "test contract"
	prg := "test contract"
	dummyContract := &registry.SCData {
		Scid: scid,
		OwnerPubKey: ownerpub,
		Description: dscr,
		Program: prg,
	}
	req.SCList = append(req.SCList, dummyContract)
	res, _ := json.Marshal(req)
	fmt.Println(string(res))
	return utils.ToJSON(c, http.StatusOK, req.SCList)
}
