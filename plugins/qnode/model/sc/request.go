package sc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"github.com/mr-tron/base58"
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

func (rf *RequestRef) Id() *RequestId {
	return NewRequestId(rf.TxId(), rf.Index())
}

func NewRequestId(txid *HashValue, index uint16) *RequestId {
	ret := new(RequestId)
	copy(ret.Bytes()[:HashSize], txid.Bytes())
	binary.LittleEndian.PutUint16(ret.Bytes()[HashSize:HashSize+2], index)
	return ret
}

func (id *RequestId) NewRequestRef() *RequestRef {
	return NewRequestRefFromTxId(id.TransactionId(), id.Index())
}

func (id *RequestId) Bytes() []byte {
	return (*id)[:]
}

func (id *RequestId) TransactionId() *HashValue {
	var ret HashValue
	copy(ret.Bytes(), id[:HashSize])
	return &ret
}

func (id *RequestId) Index() uint16 {
	return tools.Uint16From2Bytes(id[HashSize : HashSize+2])
}

func (id *RequestId) String() string {
	return base58.Encode(id.Bytes())
}

func (id *RequestId) Short() string {
	return fmt.Sprintf("%s..[%d]", base58.Encode(id.TransactionId().Bytes()[:6]), id.Index())
}

func (id *RequestId) Shortest() string {
	return fmt.Sprintf("%s..[%d]", base58.Encode(id.TransactionId().Bytes()[:4]), id.Index())
}

func (id *RequestId) Equal(id1 *RequestId) bool {
	return bytes.Compare(id.Bytes(), id1.Bytes()) == 0
}
