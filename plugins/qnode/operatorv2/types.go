package operator

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/vm"
	"github.com/iotaledger/hive.go/logger"
	"sync"
	"time"
)

type scOperator struct {
	sync.RWMutex
	dismissed bool
	scid      *HashValue
	cfgData   *registry.ConfigData
	processor vm.Processor
	stateTx   sc.Transaction

	// peerIndex -> stateIndex -> list of req which are >= the current state index
	requestNotificationsReceived []map[uint32][]*sc.RequestId

	leaderPeerIndexList       []uint16
	currLeaderSeqIndex        int16
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time

	requests          map[HashValue]*request
	processedRequests map[HashValue]time.Duration
	inChan            chan interface{}
	comm              *messaging.CommitteeConn
	stopClock         func()
	msgCounter        int
}

// keeps stateTx of the request
type request struct {

	// id of the hash of request tx id and request block index
	reqId *HashValue

	// time when request message was received by the operator
	whenMsgReceived time.Time

	// request message as received by the operator.
	// Contains parsed SC transaction and the request block index
	reqRef *sc.RequestRef

	// index of state leader of which last notified about the request message
	// after change of the state next leader seq idx 0 must be notified and this index will be changed
	lastNotifiedLeaderOfStateIndex uint32

	// seq index of the leader last notified
	// after leader rotation next leader must be notified and this index must be uodated
	lastNotifiedLeaderSeqIndex uint16

	pushMessages                 map[HashValue][]*pushResultMsg // by result hash. Some result hashes may be from future context
	pullMessages                 map[uint16]*pullResultMsg
	ownResultCalculated          *resultCalculated       // can be nil or the record with config and stateTx equal to the current
	startedCalculation           map[HashValue]time.Time // by result hash. Flag inidcates asyn calculation started
	whenLastPushed               time.Time
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
