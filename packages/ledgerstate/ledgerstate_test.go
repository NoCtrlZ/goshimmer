package ledgerstate

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/parameter"

	"github.com/iotaledger/hive.go/objectstorage"
)

var (
	iota_          = NewColor("IOTA")
	eth            = NewColor("ETH")
	transferHash1  = NewTransferHash("TRANSFER1")
	transferHash2  = NewTransferHash("TRANSFER2")
	transferHash3  = NewTransferHash("TRANSFER3")
	transferHash4  = NewTransferHash("TRANSFER4")
	transferHash5  = NewTransferHash("TRANSFER5")
	transferHash6  = NewTransferHash("TRANSFER6")
	addressHash1   = NewAddressHash("ADDRESS1")
	addressHash2   = NewAddressHash("ADDRESS2")
	addressHash3   = NewAddressHash("ADDRESS3")
	addressHash4   = NewAddressHash("ADDRESS4")
	addressHash5   = NewAddressHash("ADDRESS5")
	addressHash6   = NewAddressHash("ADDRESS6")
	pendingReality = NewRealityId("PENDING")
)

func init() {
	if err := parameter.FetchConfig(false); err != nil {
		panic(err)
	}
}

func Benchmark(b *testing.B) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, NewColoredBalance(eth, 1337),
	)

	b.ResetTimer()

	lastTransferHash := transferHash1

	for i := 0; i < b.N; i++ {
		newTransferHash := NewTransferHash(strconv.Itoa(i))

		if err := ledgerState.BookTransfer(NewTransfer(newTransferHash).AddInput(
			NewTransferOutputReference(lastTransferHash, addressHash1),
		).AddOutput(
			addressHash1, NewColoredBalance(eth, 1337),
		)); err != nil {
			b.Error(err)
		}

		lastTransferHash = newTransferHash
	}
}

func Test(t *testing.T) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, NewColoredBalance(eth, 1337), NewColoredBalance(iota_, 1338),
	)

	ledgerState.CreateReality(pendingReality)

	transfer := NewTransfer(transferHash2).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, NewColoredBalance(iota_, 338),
	).AddOutput(
		addressHash3, NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, NewColoredBalance(iota_, 1000),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1000),
	)

	if err := ledgerState.BookTransfer(transfer); err != nil {
		t.Error(err)
	}

	if err := ledgerState.BookTransfer(NewTransfer(transferHash3).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, NewColoredBalance(iota_, 338),
	).AddOutput(
		addressHash3, NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, NewColoredBalance(iota_, 1000),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1000),
	)); err != nil {
		t.Error(err)
	}

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			fmt.Println(object.(*TransferOutput))
		})

		return true
	})
}

var transferHashCounter = 0

func generateRandomTransferHash() TransferHash {
	transferHashCounter++

	return NewTransferHash("TRANSFER" + strconv.Itoa(transferHashCounter))
}

var addressHashCounter = 0

func generateRandomAddressHash() AddressHash {
	addressHashCounter++

	return NewAddressHash("ADDRESS" + strconv.Itoa(addressHashCounter))
}

func initializeLedgerStateWithBalances(numberOfBalances int) (ledgerState *LedgerState, result []*TransferOutputReference) {
	ledgerState = NewLedgerState("testLedger").Prune()

	for i := 0; i < numberOfBalances; i++ {
		transferHash := generateRandomTransferHash()
		addressHash := generateRandomAddressHash()

		ledgerState.AddTransferOutput(transferHash, addressHash, NewColoredBalance(iota_, 1337))

		result = append(result, NewTransferOutputReference(transferHash, addressHash))
	}

	return
}

func doubleSpend(ledgerState *LedgerState, transferOutputReference *TransferOutputReference) (result []*TransferOutputReference) {
	for i := 0; i < 2; i++ {
		result = append(result, spend(ledgerState, transferOutputReference))
	}

	return
}

func spend(ledgerState *LedgerState, transferOutputReferences ...*TransferOutputReference) (result *TransferOutputReference) {
	transferHash := generateRandomTransferHash()
	addressHash := generateRandomAddressHash()

	transfer := NewTransfer(transferHash).AddOutput(
		addressHash, NewColoredBalance(iota_, uint64(len(transferOutputReferences))*1337),
	)
	for _, transferOutputReference := range transferOutputReferences {
		transfer.AddInput(transferOutputReference)
	}

	if err := ledgerState.BookTransfer(transfer); err != nil {
		panic(err)
	}

	result = NewTransferOutputReference(transferHash, addressHash)

	return
}

func spend2(ledgerState *LedgerState, transferOutputReferences ...*TransferOutputReference) (result *TransferOutputReference) {
	transferHash := generateRandomTransferHash()
	addressHash := generateRandomAddressHash()

	transfer := NewTransfer(transferHash).AddOutput(
		addressHash, NewColoredBalance(iota_, uint64(len(transferOutputReferences))*2*1337),
	)
	for _, transferOutputReference := range transferOutputReferences {
		transfer.AddInput(transferOutputReference)
	}

	if err := ledgerState.BookTransfer(transfer); err != nil {
		panic(err)
	}

	result = NewTransferOutputReference(transferHash, addressHash)

	return
}

func TestElevateAggregatedReality(t *testing.T) {
	ledgerState, transferOutputs := initializeLedgerStateWithBalances(3)

	// create 2 double spends
	doubleSpentOutputs1 := doubleSpend(ledgerState, transferOutputs[0])
	doubleSpentOutputs2 := doubleSpend(ledgerState, transferOutputs[1])
	normalSpend := spend(ledgerState, transferOutputs[2])
	_ = doubleSpend(ledgerState, normalSpend)

	// send funds from one of the double spends further
	spentInput := spend(ledgerState, doubleSpentOutputs1[1])

	// aggregate further sent funds with other reality
	outputOfAggregatedReality := spend(ledgerState, spentInput, doubleSpentOutputs2[0])

	// double spend further spend to elevate aggregated reality
	spend(ledgerState, doubleSpentOutputs1[1])

	// double spend funds of aggregated reality
	spend(ledgerState, spentInput, doubleSpentOutputs2[0])

	// spend funds of conflict in aggregated reality further
	spend2(ledgerState, outputOfAggregatedReality)

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	if err := ledgerState.GenerateRealityVisualization("realities.png"); err != nil {
		t.Error(err)
	}

	if err := NewVisualizer(ledgerState).RenderTransferOutputs("outputs.png"); err != nil {
		t.Error(err)
	}
}

func TestElevate(t *testing.T) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, NewColoredBalance(eth, 1337), NewColoredBalance(iota_, 1338),
	)

	// create first legit spend
	if err := ledgerState.BookTransfer(NewTransfer(transferHash2).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash2, NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash2, NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	// send funds further
	if err := ledgerState.BookTransfer(NewTransfer(transferHash3).AddInput(
		NewTransferOutputReference(transferHash2, addressHash2),
	).AddOutput(
		addressHash4, NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	if err := ledgerState.BookTransfer(NewTransfer(transferHash4).AddInput(
		NewTransferOutputReference(transferHash2, addressHash2),
	).AddOutput(
		addressHash4, NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	// aggregate realities
	if err := ledgerState.BookTransfer(NewTransfer(transferHash6).AddInput(
		NewTransferOutputReference(transferHash3, addressHash4),
	).AddInput(
		NewTransferOutputReference(transferHash4, addressHash4),
	).AddOutput(
		addressHash6, NewColoredBalance(iota_, 2676),
	).AddOutput(
		addressHash6, NewColoredBalance(eth, 2674),
	)); err != nil {
		t.Error(err)
	}

	// create double spend for first transfer
	if err := ledgerState.BookTransfer(NewTransfer(transferHash5).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash5, NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash5, NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			fmt.Println(object.(*TransferOutput))
		})

		return true
	})
}