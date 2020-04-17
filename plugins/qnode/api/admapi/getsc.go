package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

type Contract struct {
	Id string `json:"id"`
	Contents string `json:"contents"`
}

func GetSCData(c echo.Context) error {
	var req registry.SCId
	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}
	res, err := req.GetSC()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
