package dkgapi

import (
	"github.com/labstack/echo"
)

func HandlerGetDks(c echo.Context) error {
	 return utils.ToJSON(c, http.StatusOK, NewDKSSetReq(&req))
}
