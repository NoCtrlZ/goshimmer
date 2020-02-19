package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/labstack/echo"
	"net/http"
)

//----------------------------------------------------------
func HandlerNewAssembly(c echo.Context) error {
	var req operator.AssemblyData

	if err := c.Bind(&req); err != nil {
		return api.ToJSON(c, http.StatusOK, &api.SimpleResponse{
			Error: err.Error(),
		})
	}
	if err := req.Save(); err != nil {
		return api.ToJSON(c, http.StatusOK, &api.SimpleResponse{Error: err.Error()})
	}
	return api.ToJSON(c, http.StatusOK, &api.SimpleResponse{})
}
