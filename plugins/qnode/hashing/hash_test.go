package hashing

import (
	"fmt"
	"testing"
	"math/rand"
)

type SampleSource struct {
	seed int64
}

func (s *SampleSource) Int63() int64 {
	return s.seed
}

func (s *SampleSource) Seed(seed int64) {
	fmt.Println(s)
}

func TestHashValueFromString(t *testing.T) {
	var text = "string text"
	h, e := HashValueFromString(text)
	if e != nil {
		t.Fatalf("error occurs")
	}
	fmt.Printf("%x len = %d bytes\n", h, len(h))
}

func TestHashData(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashData(bytes)
	fmt.Printf("%x len = %d bytes\n", h, len(h))
}

func TestHashDataBlake2b(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashDataBlake2b(bytes)
	fmt.Printf("%x len = %d bytes\n", h, len(h))
}

func TestHashDataSha3(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashDataSha3(bytes)
	fmt.Printf("%x len = %d bytes\n", h, len(h))
}

func TestHashStrings(t *testing.T) {
	var str = []string{"kuku", "mumu", "zuzu", "rrrr"}
	h := HashStrings(str...)
	fmt.Printf("%x len = %d bytes\n", h[:], len(h[:]))
}

func TestRandomHash(t *testing.T) {
	var src = &SampleSource{
		seed: 1,
	}
	var rnd = rand.New(src)
	h := RandomHash(rnd)
	fmt.Printf("%x len = %d bytes\n", h[:], len(h[:]))
}

func TestHashInList(t *testing.T) {
	var seed1 = "alice"
	var seed2 = "bob"
	var seed3 = "crea"
	var seed4 = "david"
	h1, _ := HashValueFromString(seed1)
	h2, _ := HashValueFromString(seed2)
	h3, _ := HashValueFromString(seed3)
	h4, _ := HashValueFromString(seed4)
	hashArray := []*HashValue {h1, h2, h3}
	res1 := HashInList(h1, hashArray)
	if !res1 {
		t.Fatalf("failed to check")
	}
	res2 := HashInList(h4, hashArray)
	if res2 == true {
		t.Fatalf("failed to check")
	}
}
