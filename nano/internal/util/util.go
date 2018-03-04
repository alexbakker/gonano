package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

func ReverseBytes(bytes []byte) []byte {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}

func AssertReaderEOF(reader *bytes.Reader) error {
	if reader.Len() != 0 {
		return fmt.Errorf("bad data length: %d unexpected bytes", reader.Len())
	}
	return nil
}

func DialTCP(addr *net.TCPAddr, timeout time.Duration) (*net.TCPConn, error) {
	// see also: go needs generics
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return conn.(*net.TCPConn), nil
}

func MustDecodeHex(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bytes
}

func MustDecodeHex32(s string) [32]byte {
	var res [32]byte
	bytes := MustDecodeHex(s)
	copy(res[:], bytes)
	return res
}

func MustDecodeHex64(s string) [64]byte {
	var res [64]byte
	bytes := MustDecodeHex(s)
	copy(res[:], bytes)
	return res
}
