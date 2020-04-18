package admapi

import (
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

type GetScListResponse struct {
	Contracts []*registry.SCData `json:"contracts"`
	Error string 				`json:"err"`
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
	req.Contracts = append(req.Contracts, dummyContract)
	return utils.ToJSON(c, http.StatusOK, req)
}
