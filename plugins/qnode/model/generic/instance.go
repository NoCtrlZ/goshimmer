package generic

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tools"
	"io"
)

var newSignedBlock func(addr, data *hashing.HashValue) SignedBlock

func SetSignedBlockConstructor(fun func(addr, data *hashing.HashValue) SignedBlock) {
	newSignedBlock = fun
}

func NewSignedBlock(addr, data *hashing.HashValue) SignedBlock {
	return newSignedBlock(addr, data)
}

func WriteSignedBlocks(w io.Writer, blocks []SignedBlock) error {
	err := tools.WriteUint16(w, uint16(len(blocks)))
	if err != nil {
		return err
	}
	for _, b := range blocks {
		err = b.Encode().Write(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadSignedBlocks(r io.Reader) ([]SignedBlock, error) {
	var num uint16
	err := tools.ReadUint16(r, &num)
	if err != nil {
		return nil, err
	}
	ret := make([]SignedBlock, num)
	for i := range ret {
		ret[i] = NewSignedBlock(nil, nil)
		err := ret[i].Encode().Read(r)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
