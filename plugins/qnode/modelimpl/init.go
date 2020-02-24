package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func InitModelImplementation() {
	sc.SetConstructors(sc.SetConstructorsParams{
		TxConstructor:           newScTransaction,
		TxParser:                newFromValueTx,
		StateBlockConstructor:   newStateBlock,
		RequestBlockConstructor: newRequestBock,
	})
	value.SetConstructors(value.SetConstructorsParams{
		UTXOConstructor:   newUTXOTransfer,
		InputConstructor:  newInput,
		OutputConstructor: newOutput,
		TxConstructor:     newValueTx,
		ParseConstructor:  parseValueTx,
	})
}
