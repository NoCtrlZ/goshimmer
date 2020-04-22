package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/labstack/echo"
	"net/http"
)

type GetAllDKSResponse struct {
	DKSs map[string][]*tcrypto.DKShare
}

type GetDKSResponse struct {
	DKSs []*tcrypto.DKShare `json:"pri_shares"`
	Err  string             `json:"err"`
}

func HandlerGetDks(c echo.Context) error {
	req := GetDKSSetReq()
	return utils.ToJSON(c, http.StatusOK, req)
}

func GetDKSSetReq() *GetDKSResponse {
	dkslist, ok := registry.GetAllDKShare()
	if !ok {
		return &GetDKSResponse{Err: "fail to get dks"}
	}
	resp := GetDKSResponse{
		DKSs: dkslist,
	}
	return &resp
}
