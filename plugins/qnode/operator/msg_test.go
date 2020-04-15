package operator

// import (
// 	"bytes"
// 	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
// 	"testing"
// )

// var amblyIdStr = "assembly1"
// var reqIdsStr = []string{"req1", "req2", "req3", "req4"}

// var result = "Hello, world 2020!"

// func TestNotifyRequestMsg(t *testing.T) {
// 	reqsIds := make([]*HashValue, 0)
// 	for i := range reqIdsStr {
// 		reqsIds = append(reqsIds, HashStrings(reqIdsStr[i]))
// 	}
// 	m1 := &NotifyRequestMsg{
// 		SenderIndex:     3,
// 		ReceivedRequest: reqsIds[2],
// 		OldestRequests:  reqsIds,
// 	}
// 	var buf bytes.Buffer
// 	encodeNotifyRequestMsg(m1, 4, &buf)
// 	m2, err := decodeNotifyRequestMsg(buf.Bytes(), 3, 4)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	if !eq1(m1, m2) {
// 		t.Errorf("not equal")
// 	}
// }

// func eq1(m1, m2 *NotifyRequestMsg) bool {
// 	if m1.SenderIndex != m2.SenderIndex {
// 		return false
// 	}
// 	if bytes.Compare((*m1.ReceivedRequest)[:], (*m2.ReceivedRequest)[:]) != 0 {
// 		return false
// 	}
// 	if len(m1.OldestRequests) != len(m2.OldestRequests) {
// 		return false
// 	}
// 	for i := range m1.OldestRequests {
// 		if bytes.Compare((*m1.OldestRequests[i])[:], (*m2.OldestRequests[i])[:]) != 0 {
// 			return false
// 		}
// 	}
// 	return true
// }

// func TestResultHashMsg(t *testing.T) {
// 	reqsIds := make([]*HashValue, 0)
// 	for i := range reqIdsStr {
// 		reqsIds = append(reqsIds, HashStrings(reqIdsStr[i]))
// 	}
// 	resultHash := HashStrings(result)
// 	dummy := HashStrings("kuku")
// 	m1 := &pushResultMsg{
// 		SenderIndex:  2,
// 		RequestId:    reqsIds[2],
// 		ResultHashes: resultHash,
// 		SigShares:    dummy.Bytes(),
// 	}
// 	var buf bytes.Buffer
// 	encodePushResultMsg(m1, &buf)
// 	m2, err := decodePushResultMsg(2, buf.Bytes())
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	if !eq2(m1, m2) {
// 		t.Errorf("not equal")
// 	}
// }

// func eq2(m1, m2 *pushResultMsg) bool {
// 	if m1.SenderIndex != m2.SenderIndex {
// 		return false
// 	}
// 	if bytes.Compare(m1.ResultHashes.Slice(), m2.ResultHashes.Slice()) != 0 {
// 		return false
// 	}
// 	if bytes.Compare(m1.RequestId.Bytes(), m2.RequestId.Bytes()) != 0 {
// 		return false
// 	}
// 	if bytes.Compare(m1.SigShares, m2.SigShares) != 0 {
// 		return false
// 	}
// 	return true
// }
