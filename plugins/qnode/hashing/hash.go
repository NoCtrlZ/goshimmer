package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"math/rand"
)

const HashSize = sha256.Size

type HashValue [HashSize]byte
type HashableBytes []byte

var nilHash HashValue
var NilHash = &nilHash

type Hashable interface {
	Hash() *HashValue
}

func (hb HashableBytes) Hash() *HashValue {
	return HashData(hb)
}

func (h *HashValue) Bytes() []byte {
	return (*h)[:]
}

func (h *HashValue) String() string {
	return hex.EncodeToString(h[:])
}

func (h *HashValue) Short() string {
	return hex.EncodeToString((*h)[:6]) + ".."
}

func (h *HashValue) Shortest() string {
	return hex.EncodeToString((*h)[:4])
}

func (h *HashValue) Equal(h1 *HashValue) bool {
	if h == h1 {
		return true
	}
	return *h == *h1
}

func (h *HashValue) Clone() *HashValue {
	var ret HashValue
	copy(ret[:], h.Bytes())
	return &ret
}

func (h *HashValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h *HashValue) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	ret, err := HashValueFromString(s)
	if err != nil {
		return err
	}
	copy(h.Bytes(), ret.Bytes())
	return nil
}

func HashValueFromString(s string) (*HashValue, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != HashSize {
		return nil, errors.New("wrong hex encoded string. Can't convert to HashValue")
	}
	var ret HashValue
	copy(ret.Bytes(), b)
	return &ret, nil
}

func HashData(data ...[]byte) *HashValue {
	h := sha3.New256()
	for _, d := range data {
		h.Write(d)
	}
	var ret HashValue
	copy(ret[:], h.Sum(nil))
	return &ret
}

func HashHashes(hash ...*HashValue) *HashValue {
	slices := make([][]byte, len(hash))
	for i := range hash {
		slices[i] = hash[i].Bytes()
	}
	return HashData(slices...)
}

func HashStrings(str ...string) *HashValue {
	tarr := make([][]byte, len(str))
	for i, s := range str {
		tarr[i] = []byte(s)
	}
	return HashData(tarr...)
}

func RandomHash(rnd *rand.Rand) *HashValue {
	s := ""
	if rnd == nil {
		s = fmt.Sprintf("%d", rand.Int())
	} else {
		s = fmt.Sprintf("%d", rnd.Int())
	}
	return HashStrings(s, s, s)
}

func HashInList(h *HashValue, list []*HashValue) bool {
	for _, h1 := range list {
		if h.Equal(h1) {
			return true
		}
	}
	return false
}
