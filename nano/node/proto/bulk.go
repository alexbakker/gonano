package proto

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
)

type BulkPullMode byte

const (
	BulkPullModeList BulkPullMode = iota
	BulkPullModeChecksum
)

type BulkPullPacket struct {
	Address nano.Address
	Hash    block.Hash
}

type BulkPullBlocksPacket struct {
	Min   block.Hash
	Max   block.Hash
	Mode  BulkPullMode
	Count uint32
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *BulkPullPacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(s.Address[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(s.Hash[:]); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *BulkPullPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	if _, err := reader.Read(s.Address[:]); err != nil {
		return err
	}

	if _, err := reader.Read(s.Hash[:]); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}

func (s *BulkPullPacket) ID() byte {
	return idPacketBulkPull
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *BulkPullBlocksPacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(s.Min[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(s.Max[:]); err != nil {
		return nil, err
	}

	if err = buf.WriteByte(byte(s.Mode)); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, s.Count); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *BulkPullBlocksPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(s.Min[:]); err != nil {
		return err
	}

	if _, err = reader.Read(s.Max[:]); err != nil {
		return err
	}

	mode, err := reader.ReadByte()
	if err != nil {
		return err
	}
	s.Mode = BulkPullMode(mode)

	if err = binary.Read(reader, binary.LittleEndian, &s.Count); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}

func (s *BulkPullBlocksPacket) ID() byte {
	return idPacketBulkPullBlocks
}
