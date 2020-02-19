package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
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
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, NewConfigReq(&req))
}

type NewConfigRequest struct {
	Id            *hashing.HashValue   `json:"id"`
	AssemblyId    *hashing.HashValue   `json:"assembly_id"`
	Accounts      []*hashing.HashValue `json:"accounts"`
	OperatorAddrs []*operator.PortAddr `json:"operator_addrs"`
}

func NewConfigReq(req *NewConfigRequest) *utils.SimpleResponse {
	ok, err := operator.ExistConfig(req.AssemblyId, req.Id)
	if err != nil {
		return &utils.SimpleResponse{Error: err.Error()}
	}
	if ok {
		return &utils.SimpleResponse{Error: "duplicated configuration"}
	}
	for _, addr := range req.Accounts {
		ks, err := tcrypto.LoadDKShare(req.AssemblyId, addr, true)
		if err != nil {
			return &utils.SimpleResponse{Error: err.Error()}
		}
		if int(ks.N) != len(req.OperatorAddrs) {
			return &utils.SimpleResponse{Error: "number of operators inconsistent with distributed key share parameters"}
		}
	}

	if !differentAddresses(req.OperatorAddrs) {
		return &utils.SimpleResponse{Error: "addresses of operators must all be different"}
	}

	cfgRecord := operator.ConfigData{
		ConfigId:          req.Id,
		AssemblyId:        req.AssemblyId,
		Created:           time.Now().UnixNano(),
		OperatorAddresses: req.OperatorAddrs,
		Accounts:          req.Accounts,
	}
	err = cfgRecord.Save()
	if err != nil {
		return &utils.SimpleResponse{Error: err.Error()}
	}
	return &utils.SimpleResponse{}
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
