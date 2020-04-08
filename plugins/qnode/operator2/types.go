package operator2

import (
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/messaging"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
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

	scid    *HashValue
	cfgData *registry.ConfigData
	// VM
	processor vm.Processor
	// current state transaction
	stateTx sc.Transaction

	requests          map[sc.RequestId]*request
	processedRequests map[sc.RequestId]time.Duration
	inChan            chan interface{}
	comm              *messaging.CommitteeConn
	stopClock         func()
	msgCounter        int

	// peerIndex -> currState [0], nextState [1] -> list of req which are >= the current state index
	requestNotificationsReceived [][2][]*sc.RequestId

	// request processing state
	// peers, sorted according to the current state hash
	// is used as leader's rotation round robin
	leaderPeerIndexList []uint16
	// current position in the round robin.
	// it is increased modulo committee size whenever leader rotates
	currLeaderSeqIndex uint16
	// my index in the leaderPeerIndexList
	myLeaderSeqIndex uint16
	// next leader rotation deadline. Normally is is set according to the timeout
	// for the leader to finalize and confirm state update
	// if deadline is set and time is due the event 'rotateLeader' is posted
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time
	// states of requests being processed: as leader and as subordinate
	requestToProcess [2][]processingStatus
}

type processingStatus struct {
	// flag is true when operator becomes leader for the state
	leader bool
	// != nil when calculations started
	req *request
	// request, selected by the peer to process
	reqId *sc.RequestId
	// timestamp proposed by the leader
	ts time.Time
	// own calculated result
	ownResult sc.Transaction
	// received from other peers as partially signed result
	MasterDataHash *HashValue
	SigBlocks      []generic.SignedBlock
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
