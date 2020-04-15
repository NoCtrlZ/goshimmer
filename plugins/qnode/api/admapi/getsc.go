package admapi

import (
	"github.com/labstack/echo"
	"fmt"
)

func GetSCData(c echo.Context) {
	fmt.Println(c)
}
