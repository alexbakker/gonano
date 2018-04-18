package block

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano"
)

type Vote struct {
	Address   nano.Address
	Signature Signature
	Sequence  uint64
	Block     Block
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (v *Vote) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(v.Address[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(v.Signature[:]); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, v.Sequence); err != nil {
		return nil, err
	}

	blockBytes, err := v.Block.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(blockBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (v *Vote) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	if _, err := reader.Read(v.Address[:]); err != nil {
		return err
	}

	if _, err := reader.Read(v.Signature[:]); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &v.Sequence); err != nil {
		return err
	}

	blockBytes := make([]byte, reader.Len())
	if _, err := reader.Read(blockBytes); err != nil {
		return err
	}

	return v.Block.UnmarshalBinary(blockBytes)
}
