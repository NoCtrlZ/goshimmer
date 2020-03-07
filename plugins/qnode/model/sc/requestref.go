package sc

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

func NewRequestRefFromTxId(reqTxId *HashValue, reqIdx uint16) *RequestRef {
	return &RequestRef{
		reqTxId:      reqTxId,
		requestIndex: reqIdx,
	}
}

func NewRequestRefFromTx(tx Transaction, reqIdx uint16) (*RequestRef, error) {
	if int(reqIdx) >= len(tx.Requests()) {
		return nil, fmt.Errorf("wrong request index")
	}
	return &RequestRef{
		tx:           tx,
		requestIndex: reqIdx,
	}, nil
}

func (rf *RequestRef) RequestBlock() Request {
	return rf.Tx().Requests()[rf.requestIndex]
}

func (rf *RequestRef) Tx() Transaction {
	if rf.tx != nil {
		return rf.tx
	}
	if rf.reqTxId == nil {
		panic("rf.reqTxId == nil")
	}
	vtx, ok := value.GetByTransactionId(rf.reqTxId)
	if !ok {
		panic("can't find transaction")
	}
	tx, err := ParseTransaction(vtx)
	if err != nil {
		panic(err)
	}
	rf.tx = tx
	return rf.tx
}

func (rf *RequestRef) Index() uint16 {
	return rf.requestIndex
}

func (rf *RequestRef) TxId() *HashValue {
	if rf.reqTxId != nil {
		return rf.reqTxId
	}
	if rf.tx == nil {
		panic("rf.reqTxId == nil")
	}
	rf.reqTxId = rf.tx.Id()
	return rf.reqTxId
}

func (rf *RequestRef) Id() *HashValue {
	return RequestId(rf.TxId(), rf.Index())
}

func RequestId(txId *HashValue, reqIndex uint16) *HashValue {
	return HashData(txId.Bytes(), tools.Uint16To2Bytes(reqIndex))
}
