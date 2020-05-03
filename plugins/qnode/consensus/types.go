package consensus

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/committee"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/sctransaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/state"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/iotaledger/hive.go/logger"
	"time"
)

type Operator struct {
	committee     committee.Committee
	dkshare       *tcrypto.DKShare
	stateTx       *sctransaction.Transaction
	variableState state.VariableState
	// VM
	processor committee.Processor

	requests          map[sctransaction.RequestId]*request
	processedRequests map[sctransaction.RequestId]time.Duration

	requestNotificationsCurrentState []*requestNotification
	requestNotificationsNextState    []*requestNotification

	leaderPeerIndexList       []uint16
	currLeaderSeqIndex        uint16
	myLeaderSeqIndex          uint16
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time
	// states of requests being processed: as leader and as subordinate

	leaderStatus             *leaderStatus
	currentStateCompRequests []*computationRequest
	nextStateCompRequests    []*computationRequest
}

type requestNotification struct {
	reqId     *sctransaction.RequestId
	peerIndex uint16
}

type leaderStatus struct {
	req          *request
	ts           time.Time
	resultTx     *sctransaction.Transaction
	finalized    bool
	signedHashes []signedHash
}

type signedHash struct {
	MasterDataHash *hashing.HashValue
	//SigBlocks      []generic.SignedBlock
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
	reqId *sctransaction.RequestId

	// request message or nil if wasn't received yet
	reqMsg *committee.RequestMsg

	// time when request message was received by the operator
	whenMsgReceived time.Time

	// request message as received by the operator.
	// Contains parsed SC transaction and the request block index
	//reqRef *sctransaction.RequestRef

	msgCounter int
	log        *logger.Logger
}

func NewOperator(committee committee.Committee, dkshare *tcrypto.DKShare) *Operator {
	return &Operator{
		committee: committee,
		dkshare:   dkshare,
	}
}

func (op *Operator) Quorum() uint16 {
	return op.dkshare.T
}
