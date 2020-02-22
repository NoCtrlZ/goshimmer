package sc

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func NewRequestRef(tx Transaction, reqIdx uint16) (*RequestRef, error) {
	if int(reqIdx) >= len(tx.Requests()) {
		return nil, fmt.Errorf("wrong request index")
	}
	return &RequestRef{
		tx:           tx,
		requestIndex: reqIdx,
	}, nil
}

func (rf *RequestRef) RequestBlock() Request {
	return rf.tx.Requests()[rf.requestIndex]
}

func (rf *RequestRef) Tx() Transaction {
	return rf.tx
}

func (rf *RequestRef) Index() uint16 {
	return rf.requestIndex
}

func (rf *RequestRef) Id() *HashValue {
	return RequestId(rf.Tx().Id(), rf.Index())
}

func RequestId(txhash *HashValue, reqIndex uint16) *HashValue {
	return HashData(txhash.Bytes(), tools.Uint16To2Bytes(reqIndex))
}
