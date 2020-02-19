package modelimpl

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

func init() {
	value.SetConstructors(value.SetConstructorsParams{
		UTXOConstructor:   newUTXOTransfer,
		InputConstructor:  newInput,
		OutputConstructor: newOutput,
		TxConstructor:     newValueTx,
		ParseConstructor:  parseValueTx,
	})
}

// implements ValueTransaction and SC transaction interfaces

type mockValueTransaction struct {
	id       *hashing.HashValue
	transfer value.UTXOTransfer
	payload  []byte
}

func newValueTx(transfer value.UTXOTransfer, payload []byte) value.Transaction {
	return &mockValueTransaction{
		transfer: transfer,
		payload:  payload,
	}
}

func (tx *mockValueTransaction) Id() *hashing.HashValue {
	if tx.id == nil {
		tx.id = hashing.HashData(tx.transfer.Id().Bytes(), tx.Payload())
	}
	return tx.id
}

func (tx *mockValueTransaction) Transfer() value.UTXOTransfer {
	return tx.transfer
}

func (tx *mockValueTransaction) Payload() []byte {
	return tx.payload
}

func (tx *mockValueTransaction) Encode() generic.Encode {
	return tx
}

// encode

func (tx *mockValueTransaction) Write(w io.Writer) error {
	err := tx.transfer.Encode().Write(w)
	if err != nil {
		return err
	}
	err = tools.WriteBytes32(w, tx.Payload())
	return err
}

func (tx *mockValueTransaction) Read(r io.Reader) error {
	transfer := value.NewUTXOTransfer()
	err := transfer.Encode().Read(r)
	if err != nil {
		return err
	}
	payload, err := tools.ReadBytes32(r)
	if err != nil {
		return err
	}
	tx.transfer = transfer
	tx.payload = payload
	return nil
}

func parseValueTx(data []byte) (value.Transaction, error) {
	rdr := bytes.NewReader(data)
	ret := &mockValueTransaction{}
	err := ret.Read(rdr)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
