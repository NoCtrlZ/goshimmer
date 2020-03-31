package messaging

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"net"
)

// extension of BufferedConnection
// handles handshake and links with peer according to handshake information

type peeredConnection struct {
	*buffconn.BufferedConnection
	peer        *qnodePeer
	handshakeOk bool
}

func newPeeredConnection(conn net.Conn, peer *qnodePeer) *peeredConnection {
	bconn := &peeredConnection{
		BufferedConnection: buffconn.NewBufferedConnection(conn),
		peer:               peer,
	}
	bconn.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		bconn.receiveData(data)
	}))
	bconn.Events.Close.Attach(events.NewClosure(func() {
		if bconn.peer != nil {
			bconn.peer.Lock()
			bconn.peer.peerconn = nil
			bconn.peer.handshakeOk = false
			if bconn.peer.stopHeartbeat != nil {
				bconn.peer.stopHeartbeat()
				bconn.peer.stopHeartbeat = nil
			}
			bconn.peer.Unlock()
		}
	}))
	return bconn
}

func (bconn *peeredConnection) receiveData(data []byte) {
	if bconn.peer != nil {
		// it is peered but maybe not handshaked (outbound)
		if bconn.peer.handshakeOk {
			// is is handshaked
			bconn.peer.receiveData(data)
			return
		}
		// not handshaked => expected handshake message
		peerAddr := string(data)
		log.Debugf("received handshake %s", peerAddr)
		if peerAddr != bconn.peer.peerPortAddr.String() {
			log.Error("closeConn the peer connection: wrong handshake message from outbound peer: expected %s got '%s'",
				bconn.peer.peerPortAddr.String(), peerAddr)
			bconn.peer.closeConn()
		} else {
			log.Infof("handshake ok with peer %s", peerAddr)
			bconn.peer.handshakeOk = true
			bconn.peer.stopHeartbeat = bconn.peer.startHeartbeat()
		}
		return
	}
	// not peered, expected handshake message (inbound)
	peerAddr := string(data)
	log.Debugf("received handshake %s", peerAddr)

	peersMutex.RLock()
	peer, ok := peers[peerAddr]
	peersMutex.RUnlock()

	if !ok || !peer.isInbound() {
		log.Errorf("inbound connection from unexpected peer %s. Closing..", peerAddr)
		_ = bconn.Close()
		return
	}
	bconn.peer = peer

	peer.Lock()
	peer.peerconn = bconn
	peer.handshakeOk = true
	peer.stopHeartbeat = peer.startHeartbeat()
	peer.Unlock()

	if err := peer.sendHandshake(); err != nil {
		log.Error("error while responding to handshake: %v. Closing connection", err)
		_ = bconn.Close()
	}
}
