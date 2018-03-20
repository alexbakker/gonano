package store

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

const (
	PendingKeySize = wallet.AddressSize + block.HashSize
)

type PendingKey [PendingKeySize]byte

type Pending struct {
	Address wallet.Address
	Amount  wallet.Balance
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *Pending) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(p.Address[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(p.Amount.Bytes(binary.BigEndian)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *Pending) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	if _, err := reader.Read(p.Address[:]); err != nil {
		return err
	}

	balance := make([]byte, wallet.BalanceSize)
	if _, err := reader.Read(balance); err != nil {
		return err
	}
	if err := p.Amount.UnmarshalBinary(balance); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}
