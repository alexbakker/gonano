package proto

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/alexbakker/gonano/nano/internal/util"
)

type KeepAlivePacket struct {
	Peers []*net.UDPAddr
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *KeepAlivePacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	writePeer := func(ip net.IP, port int) error {
		if _, err := buf.Write(ip.To16()); err != nil {
			return err
		}

		return binary.Write(buf, binary.LittleEndian, uint16(port))
	}

	for _, peer := range s.Peers {
		writePeer(peer.IP, peer.Port)
	}

	// due to a bug in the C++ implementation, we fill the list up to 8 peers
	// with unspecified ip addresses to prevent an out of bounds read
	for i := 0; i < 8-len(s.Peers); i++ {
		writePeer(net.IPv6unspecified, 0)
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *KeepAlivePacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	const ipPortSize = net.IPv6len + 2
	if len(data)%ipPortSize != 0 {
		return ErrBadLength
	}

	count := len(data) / ipPortSize
	var peers []*net.UDPAddr
	for i := 0; i < count; i++ {
		ip := make(net.IP, net.IPv6len)
		if _, err := reader.Read(ip); err != nil {
			return err
		}

		var port uint16
		if err := binary.Read(reader, binary.LittleEndian, &port); err != nil {
			return err
		}

		// don't include unspecified ip addresses
		if ip.IsUnspecified() {
			continue
		}

		peers = append(peers, &net.UDPAddr{IP: ip, Port: int(port)})
	}
	s.Peers = peers

	return util.AssertReaderEOF(reader)
}

func (s *KeepAlivePacket) ID() byte {
	return idPacketKeepAlive
}
