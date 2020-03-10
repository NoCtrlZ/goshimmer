package operator

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/fairroulette"
	"github.com/iotaledger/hive.go/logger"
	"math/rand"
	"net"
	"sync"
	"time"
)

type AssemblyOperator struct {
	sync.RWMutex
	dismissed         bool
	assemblyId        *HashValue
	cfgData           *registry.ConfigData
	processor         vm.Processor
	stateTx           sc.Transaction
	requests          map[HashValue]*request
	processedRequests map[HashValue]time.Duration
	inChan            chan interface{}
	peers             []*net.UDPAddr
	comm              messaging.Messaging
	stopClock         func()
	msgCounter        int
	rand              *rand.Rand
}

// keeps stateTx of the request
type request struct {
	reqId                        *HashValue
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

func newFromState(tx sc.Transaction, comm messaging.Messaging) (*AssemblyOperator, error) {
	state, _ := tx.State()
	oa, op := comm.GetOwnAddressAndPort()

	ret := &AssemblyOperator{
		assemblyId:        state.AssemblyId(),
		processor:         fairroulette.New(),
		requests:          make(map[HashValue]*request),
		processedRequests: make(map[HashValue]time.Duration),
		stateTx:           tx,
		inChan:            make(chan interface{}, inChanBufLen),
		comm:              comm,
	}

	iAmParticipant, err := ret.configure(state.Config().Id(), oa, op)

	if err != nil {
		return nil, err
	}
	if !iAmParticipant {
		return nil, nil
	}
	ret.startRoutines()
	return ret, nil
}

func (op *AssemblyOperator) configure(cfgId *HashValue, ownAddr string, ownPort int) (bool, error) {
	cfg, err := registry.LoadConfig(op.assemblyId, cfgId)
	if err != nil {
		return false, err
	}
	peers, err := makePeers(cfg.NodeAddresses, cfg.Index, ownAddr, ownPort)
	if err != nil {
		return false, err
	}
	if peers == nil {
		return false, nil // not participant
	}
	op.cfgData = cfg
	op.peers = peers
	op.rand = rand.New(rand.NewSource(int64(cfg.Index)))
	return true, nil
}

func makePeers(addrs []*registry.PortAddr, index uint16, ownAddr string, ownPort int) ([]*net.UDPAddr, error) {
	ret := make([]*net.UDPAddr, len(addrs))

	iAmAmongOperators := false
	for i, a := range addrs {
		addr, port := a.AdjustedIP()
		if uint16(i) == index {
			if ownAddr != addr || ownPort != port {
				return nil, fmt.Errorf("inconsistent peer index %d and network address: own %s:%d got %s:%d",
					index, ownAddr, ownPort, addr, port)
			}
			iAmAmongOperators = true
			continue
		}
		ret[i] = &net.UDPAddr{
			IP:   net.ParseIP(addr),
			Port: port,
			Zone: "",
		}
	}
	if !iAmAmongOperators {
		return nil, nil
	}
	return ret, nil
}

func (op *AssemblyOperator) assemblyQuorum() uint16 {
	return op.cfgData.T
}

func (op *AssemblyOperator) assemblySize() uint16 {
	return op.cfgData.N
}

func (op *AssemblyOperator) peerIndex() uint16 {
	return op.cfgData.Index
}
