package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/operator"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

func HandlerNewConfig(c echo.Context) error {
	var req NewConfigRequest

	if err := c.Bind(&req); err != nil {
		return api.ToJSON(c, http.StatusOK, &NewConfigResponse{
			Err: err.Error(),
		})
	}
	return api.ToJSON(c, http.StatusOK, NewConfigReq(&req))
}

type NewConfigRequest struct {
	AssemblyId    *hashing.HashValue   `json:"assembly_id"`
	DKShareId     *hashing.HashValue   `json:"dkshare_id"`
	Id            *hashing.HashValue   `json:"id"`
	OperatorAddrs []*operator.PortAddr `json:"operator_addrs"`
}

type NewConfigResponse struct {
	Err string `json:"err"`
}

func NewConfigReq(req *NewConfigRequest) *NewConfigResponse {
	ok, err := operator.ExistPrivateConfig(req.AssemblyId, req.Id)
	if err != nil {
		return &NewConfigResponse{Err: err.Error()}
	}
	if ok {
		return &NewConfigResponse{Err: "duplicated private configuration"}
	}
	ks, err := tcrypto.LoadDKShare(req.AssemblyId, req.DKShareId, true)
	if err != nil {
		return &NewConfigResponse{Err: err.Error()}
	}
	if int(ks.N) != len(req.OperatorAddrs) {
		return &NewConfigResponse{Err: "number of operators inconsistent with distributed key share parameters"}
	}
	if !differentAddresses(req.OperatorAddrs) {
		return &NewConfigResponse{Err: "addresses of operators must all be different"}
	}
	privCfg := operator.ConfigData{
		ConfigId:          req.Id,
		AssemblyId:        req.AssemblyId,
		Created:           time.Now().UnixNano(),
		OperatorAddresses: req.OperatorAddrs,
		DKeyIds:           []*hashing.HashValue{req.DKShareId},
	}
	err = privCfg.Save()
	if err != nil {
		return &NewConfigResponse{Err: err.Error()}
	}
	return &NewConfigResponse{}
}

func differentAddresses(addrs []*operator.PortAddr) bool {
	if len(addrs) <= 1 {
		return true
	}
	for i := 0; i < len(addrs)-1; i++ {
		for j := i + 1; j < len(addrs); j++ {
			if addrs[i].Addr == addrs[j].Addr && addrs[i].Port == addrs[j].Port {
				return false
			}
		}
	}
	return true
}
