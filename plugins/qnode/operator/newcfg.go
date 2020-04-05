package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
	"github.com/iotaledger/hive.go/logger"
	"sync"
	"time"
)

type scOperator struct {
	sync.RWMutex
	dismissed         bool
	scid              *HashValue
	cfgData           *registry.ConfigData
	processor         vm.Processor
	stateTx           sc.Transaction
	requests          map[sc.RequestId]*request
	processedRequests map[sc.RequestId]time.Duration
	inChan            chan interface{}
	comm              *messaging.CommitteeConn
	stopClock         func()
	msgCounter        int
}

// keeps stateTx of the request
type request struct {
	reqId                        *sc.RequestId
	whenMsgReceived              time.Time
	reqRef                       *sc.RequestRef
	pushMessages                 map[HashValue][]*pushResultMsg // by result hash. Some result hashes may be from future context
	pullMessages                 map[uint16]*pullResultMsg
	ownResultCalculated          *resultCalculated       // can be nil or the record with config and stateTx equal to the current
	startedCalculation           map[HashValue]time.Time // by result hash. Flag inidcates asyn calculation started
	leaderPeerIndexList          []uint16
	whenLastPushed               time.Time
	currLeaderSeqIndex           int16
	hasBeenPushedToCurrentLeader bool
	msgCounter                   int
	log                          *logger.Logger
}

type resultCalculated struct {
	res            *runtimeContext
	resultHash     *HashValue
	masterDataHash *HashValue
	// processing stateTx
	pullSent         bool
	whenLastPullSent time.Time
	// finalization
	finalized     bool
	finalizedWhen time.Time
}

const inChanBufLen = 10

func newFromState(tx sc.Transaction) (*scOperator, error) {
	state, _ := tx.State()

	ret := &scOperator{
		scid:              state.SContractId(),
		processor:         fairroulette.New(),
		requests:          make(map[sc.RequestId]*request),
		processedRequests: make(map[sc.RequestId]time.Duration),
		stateTx:           tx,
		inChan:            make(chan interface{}, inChanBufLen),
	}

	iAmParticipant, err := ret.configure(state.Config().Id())

	if err != nil {
		return nil, err
	}
	if !iAmParticipant {
		return nil, nil
	}
	ret.comm = messaging.RegisterNewOperator(ret, func(senderIndex uint16, msgType byte, msgData []byte) {
		ret.receiveMsgData(senderIndex, msgType, msgData)
	})

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
