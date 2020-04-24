package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestBasicScId(t *testing.T) {
	addr := address.RandomOfType(address.VERSION_BLS)
	color := RandomColor()
	scid := NewScId(&addr, color)

	scidstr := scid.String()
	scid1, err := ScIdFromString(scidstr)
	assert.Equal(t, err, nil)
	assert.Equal(t, scidstr, scid1.String())

	assert.Equal(t, bytes.Equal(scid.Address().Bytes(), addr[:]), true)
	assert.Equal(t, bytes.Equal(scid.Color().Bytes(), color[:]), true)
}

const (
	testAddress = "kKELws7qgMmpsufwf13CEQkRmYbCnrTg7f1qKNRgyVZ7"
	testScid    = "DsHiYnydheNLfhkc9sYPySVcEnyhxgtP4wWhKsczbnrRpYXrabwEjuej2N7bvb1qtdgGMewiWonzsD1zmLJAAXdE"
)

func TestRandScid(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)
	assert.Equal(t, addr.Version(), address.VERSION_BLS)

	scid := RandomScId(&addr)
	a := scid.Address().Bytes()
	assert.Equal(t, bytes.Equal(a, addr[:]), true)
	//t.Logf("scid = %s", scid.String())

	scid1, err := ScIdFromString(scid.String())
	assert.Equal(t, err, nil)
	assert.Equal(t, scid.Equal(scid1), true)
}

func TestTransactionStateBlock1(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputId(addr, valuetransaction.RandomId())
	txb.AddInputs(o1)
	bal := balance.New(balance.COLOR_IOTA, 1)
	txb.AddOutput(addr, bal)

	scid, _ := ScIdFromString(testScid)
	txb.AddStateBlock(scid, 42)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionStateBlock2(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputId(addr, valuetransaction.RandomId())
	txb.AddInputs(o1)
	bal := balance.New(balance.COLOR_IOTA, 1)
	txb.AddOutput(addr, bal)

	scid, _ := ScIdFromString(testScid)
	txb.AddStateBlock(scid, 42)
	txb.SetStateBlockParams(StateBlockParams{
		Timestamp:       time.Now().UnixNano(),
		RequestId:       NewRandomRequestId(2),
		StateUpdateHash: *hashing.RandomHash(nil),
	})
	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionRequestBlock(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputId(addr, valuetransaction.RandomId())
	txb.AddInputs(o1)
	bal := balance.New(balance.COLOR_IOTA, 1)
	txb.AddOutput(addr, bal)

	scid, _ := ScIdFromString(testScid)

	reqBlk := NewRequestBlock(scid)
	txb.AddRequestBlock(reqBlk)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionMultiBlocks(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputId(addr, valuetransaction.RandomId())
	txb.AddInputs(o1)
	bal := balance.New(balance.COLOR_IOTA, 1)
	txb.AddOutput(addr, bal)

	scid, _ := ScIdFromString(testScid)

	txb.AddStateBlock(scid, 42)
	txb.SetStateBlockParams(StateBlockParams{
		Timestamp:       time.Now().UnixNano(),
		RequestId:       NewRandomRequestId(2),
		StateUpdateHash: *hashing.RandomHash(nil),
	})

	reqBlk := NewRequestBlock(scid)
	txb.AddRequestBlock(reqBlk)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}
