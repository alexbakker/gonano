package proto

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

type FrontierReqPacket struct {
	StartAddress wallet.Address
	Age          uint32
	Count        uint32
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *FrontierReqPacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(s.StartAddress); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, s.Age); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, s.Count); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *FrontierReqPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	s.StartAddress = make([]byte, wallet.AddressSize)
	if _, err := reader.Read(s.StartAddress); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &s.Age); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &s.Count); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}

func (s *FrontierReqPacket) ID() byte {
	return idPacketFrontierReq
}
