package operator2

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
	"time"
)

const inChanBufLen = 10

func newFromState(tx sc.Transaction) (*scOperator, error) {
	state, _ := tx.State()

	ret := &scOperator{
		scid:              state.SContractId(),
		processor:         fairroulette.New(),
		requests:          make(map[sc.RequestId]*request),
		processedRequests: make(map[sc.RequestId]time.Duration),
		inChan:            make(chan interface{}, inChanBufLen),
	}

	iAmParticipant, err := ret.configure(state.Config().Id())

	if err != nil {
		return nil, err
	}
	if !iAmParticipant {
		return nil, nil
	}
	ret.requestNotificationsCurrentState = make([]*requestNotification, 0, ret.CommitteeSize())
	ret.requestNotificationsNextState = make([]*requestNotification, 0, ret.CommitteeSize())

	ret.comm = messaging.RegisterNewOperator(ret, func(senderIndex uint16, msgType byte, msgData []byte, ts time.Time) {
		ret.receiveMsgData(senderIndex, msgType, msgData, ts)
	})
	ret.setNewState(tx)
	ret.startRoutines()
	return ret, nil
}

func (op *scOperator) configure(cfgId *HashValue) (bool, error) {
	var err error
	op.cfgData, err = registry.LoadConfig(op.scid, cfgId)
	if err != nil {
		return false, err
	}
	if op.cfgData.Index >= op.cfgData.N || int(op.cfgData.N) != len(op.cfgData.NodeAddresses) {
		return false, fmt.Errorf("inconsistent config data scid: %s cfg id: %s",
			op.cfgData.AssemblyId, op.cfgData.ConfigId)
	}
	return messaging.OwnPortAddr().String() == op.cfgData.NodeAddresses[op.cfgData.Index].String(), nil
}
