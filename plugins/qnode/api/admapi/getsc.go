package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/labstack/echo"
	"net/http"
)

func GetSCData(c echo.Context) error {
	return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{})
}
