package main

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/value"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/rep"
)

type remoteTxDb struct {
	sock mangos.Socket
}

const (
	TXDB_REQ_PUT         = byte(1)
	TXDB_REQ_GET_BY_TXID = byte(2)
	TXDB_REQ_GET_BY_TRID = byte(3)
)

func (rdb *remoteTxDb) PutTransaction(tx value.Transaction) bool {
	var buf bytes.Buffer
	buf.WriteByte(TXDB_REQ_PUT)
	_ = tx.Encode().Write(&buf)
	// send to host
	// don wait for response
}

func (rdb *remoteTxDb) GetByTransactionId(id *hashing.HashValue) (value.Transaction, bool) {
	panic("implement me")
}

func (rdb *remoteTxDb) GetByTransferId(id *hashing.HashValue) (value.Transaction, bool) {
	panic("implement me")
}

func (rdb *remoteTxDb) GetSpendingTransaction(outputRefId *hashing.HashValue) (value.Transaction, bool) {
	panic("implement me")
}

func runTxDbServer() {
	var sock mangos.Socket
	var err error
	var msg []byte
	if sock, err = rep.NewSocket(); err != nil {
	}
	if err = sock.Listen(url); err != nil {
	}
	for {
		// Could also use sock.RecvMsg to get header
		msg, err = sock.Recv()
		if string(msg) == "DATE" { // no need to terminate
			fmt.Println("NODE0: RECEIVED DATE REQUEST")
			d := date()
			fmt.Printf("NODE0: SENDING DATE %s\n", d)
			err = sock.Send([]byte(d))
			if err != nil {
				die("can't send reply: %s", err.Error())
			}
		}
	}
}
