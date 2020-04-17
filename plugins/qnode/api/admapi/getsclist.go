package admapi

import (
	"time"
	"github.com/labstack/echo"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/hive.go/database"
	"net/http"
)

type GetScListResponse struct {
	Contracts []*database.Entry `json:"contracts"`
	Error string 				`json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	key := []byte{1, 2, 3, 4, 5}
	value := []byte{6, 7, 8, 9, 10}
	duration := 2
	dummyEntry := &database.Entry {
		Key: key,
		Value: value,
		Meta: 11,
		TTL: time.Duration(duration),
	}
	return utils.ToJSON(c, http.StatusOK, dummyEntry)
}
