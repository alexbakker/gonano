package store

import (
	"bytes"
	"encoding/binary"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

type AddressInfo struct {
	HeadBlock block.Hash
	RepBlock  block.Hash
	OpenBlock block.Hash
	Balance   wallet.Balance
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (i *AddressInfo) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(i.HeadBlock[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(i.RepBlock[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(i.OpenBlock[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(i.Balance.Bytes(binary.BigEndian)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (i *AddressInfo) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(i.HeadBlock[:]); err != nil {
		return err
	}

	if _, err = reader.Read(i.RepBlock[:]); err != nil {
		return err
	}

	if _, err = reader.Read(i.OpenBlock[:]); err != nil {
		return err
	}

	balance := make([]byte, wallet.BalanceSize)
	if _, err = reader.Read(balance); err != nil {
		return err
	}
	if err = i.Balance.UnmarshalBinary(balance); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}
