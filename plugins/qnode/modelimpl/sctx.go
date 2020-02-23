package modelimpl

import (
	"bytes"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
)

// implements ValueTransaction and SC transaction interfaces

func init() {
	sc.SetConstructors(sc.SetConstructorsParams{
		TxConstructor:           newScTransaction,
		TxParser:                newFromValueTx,
		StateBlockConstructor:   newStateBlock,
		RequestBlockConstructor: newRequestBock,
	})
}

type mockScTransaction struct {
	id         *HashValue
	vtx        value.Transaction
	transfer   value.UTXOTransfer
	stateBlock sc.State
	reqBlocks  []sc.Request
}

func newScTransaction() sc.Transaction {
	return &mockScTransaction{
		transfer: value.NewUTXOTransfer(),
	}
}

func (tx *mockScTransaction) SetState(state sc.State) {
	tx.stateBlock = state
}

func (tx *mockScTransaction) AddRequest(req sc.Request) {
	if tx.reqBlocks == nil {
		tx.reqBlocks = make([]sc.Request, 0)
	}
	tx.reqBlocks = append(tx.reqBlocks, req)
}

func (tx *mockScTransaction) Equal(tx1 sc.Transaction) bool {
	if sc.Transaction(tx) == tx1 {
		return true
	}
	return tx.Id().Equal(tx1.Id())
}

func (tx *mockScTransaction) Transfer() value.UTXOTransfer {
	return tx.transfer
}

func (tx *mockScTransaction) Id() *HashValue {
	if tx.id != nil {
		return tx.id
	}
	vtx, err := tx.ValueTx()
	if err != nil {
		return nil
	}
	tx.id = vtx.Id()
	return tx.id
}

func (tx *mockScTransaction) ShortStr() string {
	return tx.Id().Short()
}

func (tx *mockScTransaction) Signatures() []generic.SignedBlock {
	trsigs := tx.Transfer().InputSignatures()
	ret := make([]generic.SignedBlock, 0, len(trsigs))
	for _, sigblk := range trsigs {
		ret = append(ret, sigblk)
	}
	return ret
}

func (tx *mockScTransaction) MasterDataHash() *HashValue {
	var buf bytes.Buffer
	buf.Write(tx.transfer.DataHash().Bytes())
	_ = tx.stateBlock.Encode().Write(&buf)
	for _, req := range tx.Requests() {
		_ = req.Encode().Write(&buf)
	}
	return HashData(buf.Bytes())
}

func (tx *mockScTransaction) State() (sc.State, bool) {
	return tx.stateBlock, tx.stateBlock != nil
}

func (tx *mockScTransaction) MustState() sc.State {
	if tx.stateBlock == nil {
		panic("MustState: not a state update")
	}
	return tx.stateBlock
}

func (tx *mockScTransaction) Requests() []sc.Request {
	return tx.reqBlocks
}

func (tx *mockScTransaction) ValueTx() (value.Transaction, error) {
	if tx.vtx != nil {
		return tx.vtx, nil
	}
	var buf bytes.Buffer
	hasState := tx.stateBlock != nil
	err := tools.WriteBoolByte(&buf, hasState)
	if err != nil {
		return nil, err
	}
	if hasState {
		err = tx.stateBlock.Encode().Write(&buf)
		if err != nil {
			return nil, err
		}
	}
	err = tools.WriteUint16(&buf, uint16(len(tx.reqBlocks)))
	if err != nil {
		return nil, err
	}
	if len(tx.reqBlocks) != 0 {
		for _, req := range tx.reqBlocks {
			err = req.Encode().Write(&buf)
			if err != nil {
				return nil, err
			}
		}
	}
	tx.vtx = value.NewTransaction(tx.transfer, buf.Bytes())
	return tx.vtx, nil
}

func newFromValueTx(vtx value.Transaction) (sc.Transaction, error) {
	tx := &mockScTransaction{}
	tx.transfer = vtx.Transfer()
	buf := bytes.NewReader(vtx.Payload())
	var hasState bool
	err := tools.ReadBoolByte(buf, &hasState)
	if err != nil {
		return nil, err
	}
	if hasState {
		tx.stateBlock = &mockStateBlock{}
		err = tx.stateBlock.Encode().Read(buf)
		if err != nil {
			return nil, err
		}
	} else {
		tx.stateBlock = nil
	}
	var num uint16
	err = tools.ReadUint16(buf, &num)
	if err != nil {
		return nil, err
	}
	tx.reqBlocks = make([]sc.Request, num)
	for i := range tx.reqBlocks {
		tx.reqBlocks[i] = &mockRequestBlock{}
		err = tx.reqBlocks[i].Encode().Read(buf)
		if err != nil {
			return nil, err
		}
	}
	return tx, nil
}
