package proto

import (
	"bytes"
)

type HandshakePacket struct {
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *HandshakePacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *HandshakePacket) UnmarshalBinary(data []byte) error {
	//reader := bytes.NewReader(data)
	return nil
}

func (s *HandshakePacket) ID() byte {
	return idPacketKeepAlive
}
