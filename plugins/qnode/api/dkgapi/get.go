package dkgapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
)

type GetAllDKSResponse struct {
	DKSs map[string][]*tcrypto.DKShare
}

type GetDKSResponse struct {
	DKSs []*tcrypto.DKShare `json:"pri_shares"`
	Err  string             `json:"err"`
}
