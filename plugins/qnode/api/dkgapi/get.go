package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/labstack/echo"
	"net/http"
	"fmt"
	"encoding/json"
)

type GetAllDKSResponse struct {
	DKSs map[string][]*tcrypto.DKShare
}

type GetDKSResponse struct {
	DKSs []*tcrypto.DKShare `json:"pri_shares"`
	Err  string             `json:"err"`
}

func HandlerGetDks(c echo.Context) error {
	return utils.ToJSON(c, http.StatusOK, GetDKSSetReq())
}

func GetDKSSetReq() *GetDKSResponse {
	dkslist, ok := registry.GetAllDKShare()
	if !ok {
		return &GetDKSResponse{Err: "fail to get dks"}
	}
	resp := GetDKSResponse{
		DKSs: dkslist,
	}
	res2B, _ := json.Marshal(resp)
	fmt.Println("here is get dks")
	fmt.Println(string(res2B))
	return &resp
}
