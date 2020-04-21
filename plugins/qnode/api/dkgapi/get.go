package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

type GetDKSResponse struct {
	DKSs map[string]*tcrypto.DKShare
}

func HandlerGetDks(c echo.Context) error {
	 return utils.ToJSON(c, http.StatusOK, GetDKSResponse)
}
