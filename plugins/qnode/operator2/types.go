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

	requestNotificationsCurrentState []*requestNotification
	requestNotificationsNextState    []*requestNotification

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

	leaderStatus             *leaderStatus
	currentStateCompRequests []*computationRequest
	nextStateCompRequests    []*computationRequest
}

type requestNotification struct {
	reqId     *sc.RequestId
	peerIndex uint16
}

type leaderStatus struct {
	req          *request
	ts           time.Time
	resultTx     sc.Transaction
	finalized    bool
	signedHashes []signedHash
}

type signedHash struct {
	MasterDataHash *HashValue
	SigBlocks      []generic.SignedBlock
}

type computationRequest struct {
	ts              time.Time
	leaderPeerIndex uint16
	req             *request
	processed       bool
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
