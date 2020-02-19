package operator

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	. "github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm/vmimpl"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"sync"
	"time"
)

type AssemblyOperator struct {
	sync.RWMutex
	dismissed            bool
	assemblyId           *HashValue
	size                 uint16
	quorum               uint16
	index                uint16
	processor            vm.Processor
	stateTx              sc.Transaction
	requests             map[HashValue]*request
	inChan               chan interface{}
	keyShares            map[HashValue]*DKShare
	peers                []*net.UDPAddr
	comm                 messaging.Messaging
	stopClock            func()
	clockTickPeriod      time.Duration
	clockTickPeriodMutex *sync.RWMutex
	msgCounter           int
	processedCounter     int
	rand                 *rand.Rand
}

// keeps stateTx of the request
type request struct {
	reqId                        *HashValue
	whenMsgReceived              time.Time
	msgTx                        sc.Transaction
	msgIndex                     uint16
	receivedResultHashes         map[HashValue][]*pushResultMsg // by result hash. Some result hashes may be from future context
	ownResultCalculated          *resultCalculatedIntern        // can be nil or the record with config and stateTx equal to the current
	pullMessages                 map[uint16]*pullResultMsg
	startedCalculation           map[HashValue]time.Time // by result hash. Flag inidcates asyn calculation started
	leaderPeerIndexList          []uint16
	whenLastPushed               time.Time
	currLeaderSeqIndex           int16
	hasBeenPushedToCurrentLeader bool
	msgCounter                   int
}

const inChanBufLen = 10

func NewFromState(tx sc.Transaction, comm messaging.Messaging) (*AssemblyOperator, error) {
	state, _ := tx.State()
	oa, op := comm.GetOwnAddressAndPort()

	allCfgData, err := loadAllConfigData(state.AssemblyId(), state.ConfigId(), oa, op)
	if err != nil {
		return nil, err
	}
	if !allCfgData.iAmParticipant {
		return nil, nil
	}
	ret := &AssemblyOperator{
		assemblyId:           state.AssemblyId(),
		size:                 allCfgData.assemblySize,
		quorum:               allCfgData.quorum,
		index:                allCfgData.index,
		processor:            vmimpl.New(),
		requests:             make(map[HashValue]*request),
		stateTx:              tx,
		keyShares:            allCfgData.dkshares,
		inChan:               make(chan interface{}, inChanBufLen),
		peers:                allCfgData.peers,
		comm:                 comm,
		clockTickPeriodMutex: &sync.RWMutex{},
		clockTickPeriod:      clockTickPeriod,
		rand:                 rand.New(rand.NewSource(int64(allCfgData.index))),
	}
	ret.startRoutines()
	return ret, nil
}

type loadAllConfigDataResult struct {
	cfgData        *ConfigData
	dkshares       map[HashValue]*DKShare
	peers          []*net.UDPAddr
	assemblySize   uint16
	quorum         uint16
	index          uint16
	iAmParticipant bool
}

func loadAllConfigData(aid, cfgId *HashValue, ownAddr string, ownPort int) (*loadAllConfigDataResult, error) {
	ret := &loadAllConfigDataResult{}
	var err error
	ret.cfgData, err = LoadConfig(aid, cfgId)
	if err != nil {
		return nil, err
	}
	ret.dkshares = make(map[HashValue]*DKShare)

	for _, dksId := range ret.cfgData.DKeyIds {
		dks, err := LoadDKShare(aid, dksId, false)
		if err != nil {
			return nil, err
		}
		ret.dkshares[*dks.Address] = dks
		if ret.assemblySize != 0 && dks.N != ret.assemblySize {
			return nil, errors.New("inconsistent assembly size parameter between config data and DKShare data")
		} else {
			ret.assemblySize = dks.N
		}
		if ret.quorum != 0 && dks.T != ret.quorum {
			return nil, errors.New("inconsistent assembly quorum data parameter between config data and DKShare data")
		} else {
			ret.quorum = dks.T
		}
		if dks.Index != ret.cfgData.Index {
			return nil, errors.New("not equal indices between config data and DKShare data")
		}
	}
	ret.peers, err = makePeers(ret.cfgData.OperatorAddresses, ret.cfgData.Index, ownAddr, ownPort)
	if err != nil {
		return nil, err
	}

	if ret.peers == nil {
		return &loadAllConfigDataResult{iAmParticipant: false}, nil
	}
	ret.iAmParticipant = true
	ret.index = ret.cfgData.Index
	return ret, nil
}

func makePeers(addrs []*PortAddr, index uint16, ownAddr string, ownPort int) ([]*net.UDPAddr, error) {
	ret := make([]*net.UDPAddr, len(addrs))

	iAmAmongOperators := false
	for i, a := range addrs {
		addr, port := a.AdjustedIP()
		if uint16(i) == index {
			if ownAddr != addr || ownPort != port {
				return nil, errors.New("inconsistent operator index and network address")
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

func (op *AssemblyOperator) reconfigure(cfg *loadAllConfigDataResult) {
	op.size = cfg.assemblySize
	op.quorum = cfg.quorum
	op.index = cfg.index
	op.keyShares = cfg.dkshares
	op.peers = cfg.peers
	op.rand = rand.New(rand.NewSource(int64(cfg.index)))
}

func (op *AssemblyOperator) requiredQuorum() uint16 {
	return op.quorum
}

func (op *AssemblyOperator) assemblySize() uint16 {
	return op.size
}

func (op *AssemblyOperator) peerIndex() uint16 {
	return op.index
}

func (op *AssemblyOperator) Dismiss() {
	op.stopClock()

	op.Lock()
	op.dismissed = true
	close(op.inChan)
	op.Unlock()
}

func (op *AssemblyOperator) IsDismissed() bool {
	op.Lock()
	defer op.Unlock()
	return op.dismissed
}

func (op *AssemblyOperator) DispatchEvent(msg interface{}) {
	op.Lock()
	if !op.dismissed {
		op.inChan <- msg
	}
	op.Unlock()
}
