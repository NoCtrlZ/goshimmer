package admapi

import (
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

type GetScListResponse struct {
	registry.SCList
	Error string `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, err := registry.GetSCList()
	if err != nil {
		return err
	}
	return utils.ToJSON(c, http.StatusOK, &GetScListResponse{SCList: sclist})
}
