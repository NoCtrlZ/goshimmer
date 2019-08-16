package heartbeat

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/identity"

	"github.com/golang/protobuf/proto"

	heartbeatproto "github.com/iotaledger/goshimmer/packages/ca/heartbeat/proto"
)

func TestMarshal(t *testing.T) {
	ownNodeId := identity.GenerateRandomIdentity().Identifier

	toggledTransactions := make([]*heartbeatproto.ToggledTransaction, 1000)

	for i := 0; i < len(toggledTransactions); i++ {
		toggledTransactions[i] = &heartbeatproto.ToggledTransaction{
			TransactionId: make([]byte, 32),
			ToggleReason:  0,
		}
	}

	ownStatement := &heartbeatproto.OpinionStatement{
		NodeId:              ownNodeId,
		Time:                uint64(time.Now().Unix()),
		ToggledTransactions: toggledTransactions,
		Signature:           make([]byte, 32),
	}

	neighborStatements := make([]*heartbeatproto.OpinionStatement, 8)
	for i := 0; i < len(neighborStatements); i++ {
		neighborStatements[i] = &heartbeatproto.OpinionStatement{
			NodeId:              ownNodeId,
			Time:                uint64(time.Now().Unix()),
			ToggledTransactions: toggledTransactions,
			Signature:           make([]byte, 32),
		}
	}

	heartbeat := &heartbeatproto.HeartBeat{
		NodeId:             ownNodeId,
		OwnStatement:       ownStatement,
		NeighborStatements: neighborStatements,
		Signature:          make([]byte, 32),
	}

	serializedHeartbeat, err := proto.Marshal(heartbeat)
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(len(serializedHeartbeat))
}
