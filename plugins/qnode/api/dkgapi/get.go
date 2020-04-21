package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerGetDks(c echo.Context) error {
	 return utils.ToJSON(c, http.StatusOK, NewDKSSetReq(&req))
}
