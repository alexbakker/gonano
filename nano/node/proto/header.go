package proto

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano/internal/util"
)

const (
	HeaderSize = 8
)

type Header struct {
	Magic        [2]byte
	VersionMax   byte
	VersionUsing byte
	VersionMin   byte
	MessageType  byte
	Extensions   uint16
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *Header) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	_, err = buf.Write(s.Magic[:])
	if err != nil {
		return nil, err
	}

	fields := []byte{
		s.VersionMax,
		s.VersionUsing,
		s.VersionMin,
		s.MessageType,
	}
	if _, err = buf.Write(fields); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, s.Extensions); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Header) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(s.Magic[:]); err != nil {
		return err
	}

	fields := make([]byte, HeaderSize-len(s.Magic)-2)
	if _, err = reader.Read(fields); err != nil {
		return err
	}
	s.VersionMax = fields[0]
	s.VersionUsing = fields[1]
	s.VersionMin = fields[2]
	s.MessageType = fields[3]

	if err = binary.Read(reader, binary.LittleEndian, &s.Extensions); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}

func (s *Header) BlockType() byte {
	return byte((s.Extensions & 0x0f00) >> 8)
}

func (s *Header) SetBlockType(b byte) {
	s.Extensions &= ^uint16(0x0f00)
	s.Extensions |= (uint16(b) << 8)
}
