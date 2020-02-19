package qserver

//
//import (
//	. "github.com/iotaledger/goshimmer/plugins/qnode/glb"
//	. "github.com/iotaledger/goshimmer/plugins/qnode/operator"
//	"net"
//	"testing"
//)
//
//var amblyIdStr = "assembly1"
//var reqIdsStr = []string{"req1", "req2", "req3", "req4"}
//
//var result = "Hello, world 2020!"
//
//var epoch = "epoch"
//
//var peers = []*net.UDPAddr{
//	{IP: net.ParseIP("0.0.0.0"), Port: 4000},
//	{IP: net.ParseIP("0.0.0.0"), Port: 4001},
//	{IP: net.ParseIP("0.0.0.0"), Port: 4002},
//	{IP: net.ParseIP("0.0.0.0"), Port: 4003},
//}
//
//const senderIndex = 1
//
//var senderUDPAddr = peers[senderIndex]
//
//func init() {
//	if err := StartServer(4000); err != nil {
//		panic(err)
//	}
//}
//
//func TestReceiveUDPDataErr1(t *testing.T) {
//	reqsIds := make([]*HashValue, 0)
//	for i := range reqIdsStr {
//		reqsIds = append(reqsIds, HashStrings(reqIdsStr[i]))
//	}
//	m1 := &NotifyRequestMsg{
//		SenderIndex:     senderIndex,
//		ReceivedRequest: reqsIds[2],
//		OldestRequests:  reqsIds,
//	}
//	assemblyId := HashStrings(amblyIdStr)
//	op := &AssemblyOperator{
//		assemblyId: assemblyId,
//		Index:      3,
//		epoch: &EpochBundle{
//			N:                      4,
//			T:                      3,
//			ConfigId:               "",
//			Enabled:                true,
//			epochId:              HashStrings(epoch),
//			assemblyId:             assemblyId,
//			RequestNotificationLen: 3,
//		},
//		peers: peers,
//	}
//	ServerInstance.operators[*assemblyId] = op
//
//	buf := make([]byte, 0)
//	buf = op.encodeMsg(m1)
//	err := receiveUDPDataErr(senderUDPAddr, buf)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestReceiveUDPDataErr2(t *testing.T) {
//	m1 := &ResultHashMsg{
//		SenderIndex: senderIndex,
//		RequestId: HashStrings(reqIdsStr[0]),
//		ResultHashes:  HashStrings(result),
//	}
//	assemblyId := HashStrings(amblyIdStr)
//	op := &AssemblyOperator{
//		assemblyId: assemblyId,
//		Index:      3,
//		epoch: &EpochBundle{
//			N:                      4,
//			T:                      3,
//			ConfigId:               "",
//			Enabled:                true,
//			epochId:              HashStrings(epoch),
//			assemblyId:             assemblyId,
//			RequestNotificationLen: 3,
//		},
//		peers: peers,
//	}
//	ServerInstance.operators[*assemblyId] = op
//
//	buf := make([]byte, 0)
//	buf = op.encodeMsg(m1)
//	err := receiveUDPDataErr(senderUDPAddr, buf)
//	if err != nil {
//		t.Error(err)
//	}
//}
