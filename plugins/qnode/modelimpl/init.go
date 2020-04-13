package modelimpl

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
)

func Init() {
	sc.SetConstructors(sc.SetConstructorsParams{
		TxConstructor:           newScTransaction,
		TxParser:                newFromValueTx,
		StateBlockConstructor:   newStateBlock,
		RequestBlockConstructor: newRequestBlock,
	})
	value.SetConstructors(value.SetConstructorsParams{
		UTXOConstructor:   newUTXOTransfer,
		InputConstructor:  newInput,
		OutputConstructor: newOutput,
		TxConstructor:     newValueTx,
		ParseConstructor:  parseValueTx,
	})
}
