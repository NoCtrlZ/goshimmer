package registry

import (
	"testing"
)

func TestAjustedIP(t *testing.T) {
	portAddr1 := &PortAddr {
		Port: 8080,
		Addr: "localhost",
	}
	addr1, port1 := portAddr1.AdjustedIP()
	if addr1 != "127.0.0.1" || port1 != 8080 {
		t.Fatalf("failed to get adjusted addr or port")
	}

	portAddr2 := &PortAddr {
		Port: 7070,
		Addr: "127.0.0.2",
	}
	addr2, port2 := portAddr2.AdjustedIP()
	if addr2 != "127.0.0.2" || port2 != 7070 {
		t.Fatalf("failed to get adjusted addr or port")
	}
}

func TestString(t *testing.T) {
	portAddr := &PortAddr {
		Port: 8080,
		Addr: "localhost",
	}
	stringAddr := portAddr.String()
	if stringAddr != "localhost:8080" {
		t.Fatalf("failed to convert string")
	}
}
