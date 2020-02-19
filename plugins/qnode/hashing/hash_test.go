package hashing

import (
	"fmt"
	"testing"
)

func TestHashStrings(t *testing.T) {
	var str = []string{"kuku", "mumu", "zuzu", "rrrr"}
	h := HashStrings(str...)
	fmt.Printf("%x len = %d bytes\n", h[:], len(h[:]))
}
