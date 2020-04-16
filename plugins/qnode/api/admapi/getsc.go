package admapi

import (
	"github.com/labstack/echo"
	"net/http"
)

type Contract struct {
	Id string `json:"id"`
	Contents string `json:"contents"`
}

func GetSCData(c echo.Context) error {
	contract := &Contract{
		Id: c.Param("id"),
		Contents: "test",
	}
	return c.JSON(http.StatusOK, contract)
}
