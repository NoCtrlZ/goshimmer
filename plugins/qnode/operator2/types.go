package operator2

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
	dismissed    bool
	scid         *HashValue
	cfgData      *registry.ConfigData
	processor    vm.Processor
	stateTx      sc.Transaction
	stateChanged bool

	// peerIndex -> currState, nextState -> list of req which are >= the current state index
	requestNotificationsReceived [][2][]*sc.RequestId

	requests          map[sc.RequestId]*request
	processedRequests map[sc.RequestId]time.Duration
	inChan            chan interface{}
	comm              *messaging.CommitteeConn
	stopClock         func()
	msgCounter        int

	// request processing state
	// peers, sorted according to the current state hash
	// is used as leader's rotation round robin
	leaderPeerIndexList []uint16
	// current position in the round robin.
	// it is increased modulo committee size whenever leader rotates
	currLeaderSeqIndex uint16
	// next leader rotation deadline. Normally is is set according to the timeout
	// for the leader to finalize and confirm state update
	// if deadline is set and time is due the event 'rotateLeader' is posted
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time
	// leader part
	// when operator becomes the leader of the current state, it selects request to process
	// stores it as currentRequest, posts 'initReq' messages and start async calculation of th result
	// this part us only needed for the operator who once became the leader of the state
	// once set, currentRequest only changes after change of the state
	currentRequest *request          // can be nil
	currentResult  *resultCalculated // if not nil, it is the result of the current context
	// non-leader part
	// requests to process, received as 'initReq' messages.
	// requestToProcess[0] corresponds to the current state index
	// requestToProcess[1] corresponds to the next state index
	// 'initReq' messages with smaller and larger indices are ignored
	// each is a slice with len = size of the committee, one element per peer
	requestToProcess [2][]*requestToProcess
}

type requestToProcess struct {
	msg                   *initReqMsg
	whenReceived          time.Time
	resultBeingCalculated bool
	resultSent            bool
}

// keeps stateTx of the request
type request struct {

	// id of the hash of request tx id and request block index
	reqId *sc.RequestId

	// time when request message was received by the operator
	whenMsgReceived time.Time

	// request message as received by the operator.
	// Contains parsed SC transaction and the request block index
	reqRef *sc.RequestRef

	// index of state leader of which last notified about the request message
	// after change of the state next leader seq idx 0 must be notified and this index will be changed
	lastNotifiedLeaderOfStateIndex uint32

	// seq index of the leader last notified
	// after leader rotation next leader must be notified and this index must be updated
	lastNotifiedLeaderSeqIndex uint16

	msgCounter int
	log        *logger.Logger
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
