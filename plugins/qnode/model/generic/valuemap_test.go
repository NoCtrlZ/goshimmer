package generic

import (
	"bytes"
	"fmt"
	"testing"
)

func TestValueMap(t *testing.T) {
	m := NewFlatValueMap()
	m.SetString("kuku", "abra kadabra")
	fmt.Printf("%+v\n", m)
	var buf bytes.Buffer
	_ = m.Encode().Write(&buf)
	mb := NewFlatValueMap()
	_ = mb.Encode().Read(&buf)
	fmt.Printf("%+v\n", mb)
}
