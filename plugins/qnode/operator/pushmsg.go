package operator

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/sc"
	"github.com/pkg/errors"
)

func (op *AssemblyOperator) validatePushMessage(msg *pushResultMsg) error {
	if len(msg.SigBlocks) == 0 {
		return errors.New("push message with 0 blocks: invalid")
	}
	return sc.VerifySignedBlocks(msg.SigBlocks, op)
}

func (op *AssemblyOperator) accountNewPushMsg(msg *pushResultMsg) error {
	if err := op.validatePushMessage(msg); err != nil {
		return err
	}
	resultHash := resultHash(msg.StateIndex, msg.RequestId, msg.MasterDataHash)
	req, _ := op.requestFromId(msg.RequestId)

	if rhlst, ok := req.receivedResultHashes[*resultHash]; !ok {
		req.receivedResultHashes[*resultHash] = make([]*pushResultMsg, op.assemblySize())
	} else {
		if rhlst[msg.SenderIndex] != nil {
			if !duplicateResultHashMessages(msg, rhlst[msg.SenderIndex]) {
				return fmt.Errorf("duplicate result hash has been received")
			}
			log.Error("duplicate result hash ignored")
		}
	}
	// if duplicate, replace the previous
	req.receivedResultHashes[*resultHash][msg.SenderIndex] = msg
	return nil
}

func duplicateResultHashMessages(msg1, msg2 *pushResultMsg) bool {
	switch {
	case msg1 == msg2:
		return true
	case msg1.SenderIndex != msg2.SenderIndex:
		return false
	case !msg1.RequestId.Equal(msg2.RequestId):
		return false
	case msg1.StateIndex != msg2.StateIndex:
		return false
	case !msg1.MasterDataHash.Equal(msg2.MasterDataHash):
		return false
	case len(msg1.SigBlocks) != len(msg2.SigBlocks):
		return false
	default:
		for i := range msg1.SigBlocks {
			if !msg1.SigBlocks[i].SignedHash().Equal(msg2.SigBlocks[i].SignedHash()) {
				return false
			}
			if !msg1.SigBlocks[i].Account().Equal(msg2.SigBlocks[i].Account()) {
				return false
			}
		}
	}
	return true
}