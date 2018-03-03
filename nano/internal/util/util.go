package util

import (
	"bytes"
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
