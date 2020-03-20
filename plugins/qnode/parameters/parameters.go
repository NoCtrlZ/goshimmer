package parameters

import (
	flag "github.com/spf13/pflag"
)

const (
	QNODE_PORT          = "qnode.port"
	MOCK_TANGLE_IP_ADDR = "qnode.mockTangleIpAddr"
	MOCK_TANGLE_PORT    = "qnode.mockTanglePort"
)

func init() {
	flag.Int(QNODE_PORT, 4000, "port for committee connection")
	flag.String(MOCK_TANGLE_IP_ADDR, "127.0.0.1", "ip address for node simulator")
	flag.Int(MOCK_TANGLE_PORT, 1000, "udp port for node simulator")
}
