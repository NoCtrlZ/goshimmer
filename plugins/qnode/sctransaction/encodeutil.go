package sctransaction

import (
	"errors"
	"hash/crc32"
)

// the byte is needed for the parses to quickly recognize
// what kind of block it is: state or request
// max number of request blocks in the transaction is 127

const stateBlockMask = 0x80

func encodeMetaByte(hasState bool, numRequests byte) (byte, error) {
	if numRequests > 127 {
		return 0, errors.New("can't be more than 127 requests")
	}
	ret := numRequests
	if hasState {
		ret = ret | stateBlockMask
	}
	return ret, nil
}

func DecodeMetaByte(b byte) (bool, byte) {
	return b|stateBlockMask != 0, b & stateBlockMask
}

func mustChecksum65Bytes(data []byte) uint32 {
	if len(data) != 65 {
		panic("mustChecksum65Bytes: wrong param")
	}
	return crc32.ChecksumIEEE(data)
}
