package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

//----------------------------------------------------------
func HandlerAssemblyData(c echo.Context) error {
	var req registry.AssemblyData

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}

	if err := req.Save(); err != nil {
		log.Errorf("failed to save assembly data: %v", err)
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{Error: err.Error()})
	}
	log.Infof("assembly data saved: id = %s descr = '%s'",
		req.AssemblyId.Short(), req.Description)

	if err := registry.RefreshAssemblyData(); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{Error: err.Error()})
	}
	return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{})
}
