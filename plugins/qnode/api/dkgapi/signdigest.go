package dkgapi

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerSignDigest(c echo.Context) error {
	var req SignDigestRequest

	if err := c.Bind(&req); err != nil {
		return api.ToJSON(c, http.StatusOK, &SignDigestResponse{
			Err: err.Error(),
		})
	}
	return api.ToJSON(c, http.StatusOK, SignDigestReq(&req))
}

type SignDigestRequest struct {
	AssemblyId *hashing.HashValue `json:"assembly_id"`
	Id         *hashing.HashValue `json:"id"`
	DataDigest *hashing.HashValue `json:"data_digest"`
}

type SignDigestResponse struct {
	SigShare string `json:"sig_share"`
	Err      string `json:"err"`
}

func SignDigestReq(req *SignDigestRequest) *SignDigestResponse {
	ks, ok, err := GetDKShare(req.AssemblyId, req.Id)
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	if !ok {
		return &SignDigestResponse{Err: "unknown key share"}
	}
	if !ks.Committed {
		return &SignDigestResponse{Err: "uncommitted key set"}
	}
	signature, err := ks.SignShare(req.DataDigest.Bytes())
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	return &SignDigestResponse{
		SigShare: hex.EncodeToString(signature),
	}
}
