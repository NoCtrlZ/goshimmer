package admapi

import (
	"github.com/labstack/echo"
	"fmt"
)

func GetSCData(c echo.Context) error {
	fmt.Println(c)
}
