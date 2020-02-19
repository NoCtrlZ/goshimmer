package generic

import (
	"bytes"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
)

func Bytes(b Encode) ([]byte, error) {
	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Hash(e Encode) *hashing.HashValue {
	b, _ := Bytes(e)
	return hashing.HashData(b)
}
