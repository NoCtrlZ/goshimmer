package admapi

import (
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

type SCList []*registry.SCData

type GetScListResponse struct {
	SCList
	Error string `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	var sclist SCList
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
	sclist = append(sclist, dummyContract)
	return utils.ToJSON(c, http.StatusOK, &GetScListResponse{SCList: sclist})
}
