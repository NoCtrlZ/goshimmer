// implement smart contract transaction.
// smart contract transaction is value transaction with special payload
package sctransaction

import (
	"bytes"
	"errors"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

// Smart contract ransaction wraps value transaction
// the stateBlock and requestBlocks are parsed from the dataPayload of the value transaction
type Transaction struct {
	*valuetransaction.Transaction
	stateBlock    *StateBlock
	requestBlocks []*RequestBlock
}

func NewTransaction(vtx *valuetransaction.Transaction, stateBlock *StateBlock, requestBlocks []*RequestBlock) (*Transaction, error) {
	ret := &Transaction{
		Transaction:   vtx,
		stateBlock:    stateBlock,
		requestBlocks: requestBlocks,
	}
	scpayload, err := ret.DataScPayloadBytes()
	if err != nil {
		return nil, err
	}
	if err := vtx.SetDataPayload(scpayload); err != nil {
		return nil, err
	}
	return ret, nil
}

// parses dataPayload
func ParseValueTransaction(vtx *valuetransaction.Transaction) (*Transaction, error) {
	dataPayload := vtx.GetDataPayload()
	if !CheckScPayloadPrefix(dataPayload) {
		// pre-parsing and rejecting if clearly not and SC transaction
		return nil, errors.New("not a SC transaction")
	}
	rdr := bytes.NewReader(dataPayload)
	ret := &Transaction{Transaction: vtx}
	if err := ret.ReadDataPayload(rdr); err != nil {
		return nil, err
	}
	return ret, nil
}

func (tx *Transaction) State() (*StateBlock, bool) {
	return tx.stateBlock, tx.stateBlock != nil
}

func (tx *Transaction) MustState() *StateBlock {
	if tx.stateBlock == nil {
		panic("MustState: state block expected")
	}
	return tx.stateBlock
}

func (tx *Transaction) Requests() []*RequestBlock {
	return tx.requestBlocks
}

func (tx *Transaction) MustRequest(index uint16) *RequestBlock {
	return tx.requestBlocks[index]
}

func (tx *Transaction) DataScPayloadBytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.WriteDataPayload(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// function writes bytes of the SC transaction-specific part
func (tx *Transaction) WriteDataPayload(w io.Writer) error {
	if tx.stateBlock == nil && len(tx.requestBlocks) == 0 {
		return errors.New("can't encode empty sc transaction")
	}
	if len(tx.requestBlocks) > 127 {
		return errors.New("max number of request blocks 127 exceeded")
	}
	numRequests := byte(len(tx.requestBlocks))
	b, err := encodeMetaByte(tx.stateBlock != nil, numRequests)
	if err != nil {
		return err
	}
	if err = util.WriteByte(w, b); err != nil {
		return err
	}
	var checksum uint32
	if tx.stateBlock != nil {
		checksum = mustChecksum65Bytes(tx.stateBlock.scid.Bytes())
	} else {
		// if there's no state block, at least one request block exists
		checksum = mustChecksum65Bytes(tx.requestBlocks[0].scid.Bytes())
	}
	if err := util.WriteUint32(w, checksum); err != nil {
		return err
	}
	if tx.stateBlock != nil {
		if err := tx.stateBlock.Write(w); err != nil {
			return err
		}
	}
	for _, reqBlk := range tx.requestBlocks {
		if err := reqBlk.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) ReadDataPayload(r io.Reader) error {
	var hasState bool
	var numRequests byte
	if b, err := util.ReadByte(r); err != nil {
		return err
	} else {
		hasState, numRequests = decodeMetaByte(b)
	}
	// ignore checksum. It is only needed to check if the dataPayload can be parsed without parsing
	var checksum uint32
	if err := util.ReadUint32(r, &checksum); err != nil {
		return err
	}
	var stateBlock *StateBlock
	if hasState {
		stateBlock = &StateBlock{}
		if err := stateBlock.Read(r); err != nil {
			return err
		}
	}
	reqBlks := make([]*RequestBlock, numRequests)
	for i := range reqBlks {
		reqBlks[i] = &RequestBlock{}
		if err := reqBlks[i].Read(r); err != nil {
			return err
		}
	}
	tx.stateBlock = stateBlock
	tx.requestBlocks = reqBlks
	return nil
}
