package sctransaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"io"
)

const ScIdLength = balance.ColorLength + address.Length

// ScId is a global identifier of the smart contract
// it consist of hash of the origin transaction (a.k.a. the color of the SC)
// and address of the SC account.
// The color part is globally unique however having only the color, in order to send a request to SC
// one would need to query tangle for the origin transaction to find the address retrieve. Besides,
// origin transaction may be snapshoted.
// Another reason to have address together with color is to be able to quickly find last state of the SC:
// that would be transaction which holds the the only output of the token with the color of SC and balance
// residing in the address.
// Therefore, the ScId contains all information about SC needed to create a valid
// request to it and to find its state on the tangle.
type ScId [ScIdLength]byte

var NilScId ScId

func NewScId(color balance.Color, addr address.Address) *ScId {
	var ret ScId
	copy(ret[:balance.ColorLength], color.Bytes())
	copy(ret[balance.ColorLength:], addr[:])
	return &ret
}

func (id *ScId) Bytes() []byte {
	return id[:]
}

func (id *ScId) Color() (ret balance.Color) {
	copy(ret[:], id[:balance.ColorLength])
	return
}

func (id *ScId) Address() (ret address.Address) {
	copy(ret[:], id[balance.ColorLength:])
	return
}

func (id *ScId) ColorBytes() []byte {
	return id[:balance.ColorLength]
}

func (id *ScId) AddressBytes() []byte {
	return id[balance.ColorLength:]
}

func (id *ScId) String() string {
	return base58.Encode(id[:])
}

func (id *ScId) Short() string {
	return fmt.Sprintf("%s../%s..", id.Address().String()[:4], id.Color().String()[:4])
}

func ScIdFromString(s string) (*ScId, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ScIdLength {
		return nil, errors.New("wrong hex encoded string. Can't convert to ScId")
	}
	var ret ScId
	copy(ret[:], b)
	return &ret, nil
}

func (id *ScId) Equal(id1 *ScId) bool {
	if id == id1 {
		return true
	}
	return bytes.Equal(id.Bytes(), id1.Bytes())
}

func (id *ScId) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *ScId) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	ret, err := ScIdFromString(s)
	if err != nil {
		return err
	}
	copy(id.Bytes(), ret.Bytes())
	return nil
}

func (id *ScId) Write(w io.Writer) error {
	_, err := w.Write(id.Bytes())
	return err
}

func (id *ScId) Read(r io.Reader) error {
	n, err := r.Read(id.Bytes())
	if err != nil {
		return err
	}
	if n != ScIdLength {
		return errors.New("not enough bytes for scid")
	}
	return nil
}
