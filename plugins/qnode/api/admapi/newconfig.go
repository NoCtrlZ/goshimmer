package admapi

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerNewConfig(c echo.Context) error {
	var req registry.ConfigData

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}
	//log.Debugf("HandlerNewConfig: %+v", req)
	return utils.ToJSON(c, http.StatusOK, NewConfigReq(&req))
}

type NewConfigResponse struct {
	ConfigId *hashing.HashValue `json:"config_id"`
	Err      string             `json:"err"`
}

func NewConfigReq(req *registry.ConfigData) *NewConfigResponse {
	req.ConfigId = registry.ConfigId(req)
	err := validateConfig(req)
	if err != nil {
		return &NewConfigResponse{
			Err: err.Error(),
		}
	}
	err = registry.SaveConfig(req)
	if err != nil {
		return &NewConfigResponse{
			Err: err.Error(),
		}
	}
	log.Infow("Created new configuration", "assembly id", req.AssemblyId, "config id", req.ConfigId)
	return &NewConfigResponse{
		ConfigId: req.ConfigId,
	}
}

func validateConfig(cfg *registry.ConfigData) error {
	if len(cfg.Accounts) == 0 {
		return fmt.Errorf("0 accounts found")
	}
	if cfg.N < 4 {
		return fmt.Errorf("assembly size must be at least 4")
	}
	if cfg.T < cfg.N/2+1 {
		return fmt.Errorf("assembly quorum must be at least N/2+1 (2*N/3+1 recommended)")
	}
	if len(cfg.NodeAddresses) != int(cfg.N) {
		return fmt.Errorf("number of nodes must be equal to the size of assembly N")
	}

	ok, err := registry.ExistConfig(cfg.AssemblyId, cfg.ConfigId)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("duplicated configuration id %s", cfg.ConfigId.Short())
	}
	// check consistency with account keys

	for _, addr := range cfg.Accounts {
		ks, ok, err := registry.GetDKShare(cfg.AssemblyId, addr)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("can't find account id %s", addr.Short())
		}
		if ks.N != cfg.N || ks.T != cfg.T {
			return fmt.Errorf("inconsistent size parameters with account id %s", addr.Short())
		}
	}

	if !differentAddresses(cfg.NodeAddresses) {
		return fmt.Errorf("addresses of operator nodes must all be different")
	}
	return nil
}

func differentAddresses(addrs []*registry.PortAddr) bool {
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
