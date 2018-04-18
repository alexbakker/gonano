package block

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/internal/util"
)

const (
	idBlockInvalid byte = iota
	idBlockNotABlock
	idBlockSend
	idBlockReceive
	idBlockOpen
	idBlockChange
	idBlockState
)

var (
	ErrBadBlockType = errors.New("bad block type")
	ErrNotABlock    = errors.New("block type is not_a_block")

	blockNames = map[byte]string{
		idBlockInvalid:   "invalid",
		idBlockNotABlock: "not_a_block",
		idBlockSend:      "send",
		idBlockReceive:   "receive",
		idBlockOpen:      "open",
		idBlockChange:    "change",
		idBlockState:     "state",
	}
)

const (
	blockSizeCommon  = SignatureSize + 8
	blockSizeOpen    = blockSizeCommon + HashSize + nano.AddressSize*2
	blockSizeSend    = blockSizeCommon + HashSize + nano.AddressSize + nano.BalanceSize
	blockSizeReceive = blockSizeCommon + HashSize*2
	blockSizeChange  = blockSizeCommon + HashSize + nano.AddressSize
	blockSizeState   = blockSizeCommon + HashSize*2 + nano.AddressSize*2 + nano.BalanceSize
)

type CommonBlock struct {
	Signature Signature `json:"signature"`
	Work      Work      `json:"work"`
}

type Block interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Hash() Hash
	Root() Hash
	Signature() Signature
	Size() int
	ID() byte
	Valid() bool
}

type OpenBlock struct {
	SourceHash     Hash         `json:"source"`
	Representative nano.Address `json:"representative"`
	Address        nano.Address `json:"address"`
	Common         CommonBlock  `json:"common"`
}

type SendBlock struct {
	PreviousHash Hash         `json:"previous"`
	Destination  nano.Address `json:"destination"`
	Balance      nano.Balance `json:"balance"`
	Common       CommonBlock  `json:"common"`
}

type ReceiveBlock struct {
	PreviousHash Hash        `json:"previous"`
	SourceHash   Hash        `json:"source"`
	Common       CommonBlock `json:"common"`
}

type ChangeBlock struct {
	PreviousHash   Hash         `json:"previous"`
	Representative nano.Address `json:"representative"`
	Common         CommonBlock  `json:"common"`
}

type StateBlock struct {
	Address        nano.Address   `json:"address"`
	PreviousHash   Hash           `json:"previous"`
	Representative nano.Address   `json:"representative"`
	Balance        nano.Balance   `json:"balance"`
	Link           [HashSize]byte `json:"link"`
	Common         CommonBlock    `json:"common"`
}

func New(blockType byte) (Block, error) {
	switch blockType {
	case idBlockOpen:
		return new(OpenBlock), nil
	case idBlockSend:
		return new(SendBlock), nil
	case idBlockReceive:
		return new(ReceiveBlock), nil
	case idBlockChange:
		return new(ChangeBlock), nil
	case idBlockNotABlock:
		return nil, ErrNotABlock
	default:
		return nil, ErrBadBlockType
	}
}

func Name(id byte) string {
	return blockNames[id]
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *CommonBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.Signature[:]); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, b.Work); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *CommonBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	if _, err := reader.Read(b.Signature[:]); err != nil {
		return err
	}

	if err := binary.Read(reader, binary.LittleEndian, &b.Work); err != nil {
		return err
	}

	return util.AssertReaderEOF(reader)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *OpenBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.SourceHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Representative[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Address[:]); err != nil {
		return nil, err
	}

	commonBytes, err := b.Common.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *OpenBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.SourceHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Representative[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Address[:]); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return b.Common.UnmarshalBinary(commonBytes)
}

func (b *OpenBlock) Hash() Hash {
	return hashBytes(b.SourceHash[:], b.Representative[:], b.Address[:])
}

func (b *OpenBlock) Root() Hash {
	return b.SourceHash
}

func (b *OpenBlock) Signature() Signature {
	return b.Common.Signature
}

func (b *OpenBlock) Size() int {
	return blockSizeOpen
}

func (b *OpenBlock) ID() byte {
	return idBlockOpen
}

func (b *OpenBlock) Valid() bool {
	return b.Common.Work.Valid(Hash(b.Address))
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *SendBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.PreviousHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Destination[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Balance.Bytes(binary.BigEndian)); err != nil {
		return nil, err
	}

	commonBytes, err := b.Common.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *SendBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.PreviousHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Destination[:]); err != nil {
		return err
	}

	balance := make([]byte, nano.BalanceSize)
	if _, err = reader.Read(balance); err != nil {
		return err
	}
	if err = b.Balance.UnmarshalBinary(balance); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return b.Common.UnmarshalBinary(commonBytes)
}

func (b *SendBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.Destination[:], b.Balance.Bytes(binary.BigEndian))
}

func (b *SendBlock) Root() Hash {
	return b.PreviousHash
}

func (b *SendBlock) Signature() Signature {
	return b.Common.Signature
}

func (b *SendBlock) Size() int {
	return blockSizeSend
}

func (b *SendBlock) ID() byte {
	return idBlockSend
}

func (b *SendBlock) Valid() bool {
	return b.Common.Work.Valid(b.PreviousHash)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *ReceiveBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.PreviousHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.SourceHash[:]); err != nil {
		return nil, err
	}

	commonBytes, err := b.Common.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *ReceiveBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.PreviousHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.SourceHash[:]); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return b.Common.UnmarshalBinary(commonBytes)
}

func (b *ReceiveBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.SourceHash[:])
}

func (b *ReceiveBlock) Root() Hash {
	return b.PreviousHash
}

func (b *ReceiveBlock) Signature() Signature {
	return b.Common.Signature
}

func (b *ReceiveBlock) Size() int {
	return blockSizeReceive
}

func (b *ReceiveBlock) ID() byte {
	return idBlockReceive
}

func (b *ReceiveBlock) Valid() bool {
	return b.Common.Work.Valid(b.PreviousHash)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *ChangeBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.PreviousHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Representative[:]); err != nil {
		return nil, err
	}

	commonBytes, err := b.Common.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *ChangeBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.PreviousHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Representative[:]); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return b.Common.UnmarshalBinary(commonBytes)
}

func (b *ChangeBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.Representative[:])
}

func (b *ChangeBlock) Root() Hash {
	return b.PreviousHash
}

func (b *ChangeBlock) Signature() Signature {
	return b.Common.Signature
}

func (b *ChangeBlock) Size() int {
	return blockSizeChange
}

func (b *ChangeBlock) ID() byte {
	return idBlockChange
}

func (b *ChangeBlock) Valid() bool {
	return b.Common.Work.Valid(b.PreviousHash)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *StateBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.Address[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.PreviousHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Representative[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Balance.Bytes(binary.BigEndian)); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Link[:]); err != nil {
		return nil, err
	}

	commonBytes, err := b.Common.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *StateBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.Address[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.PreviousHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Representative[:]); err != nil {
		return err
	}

	balance := make([]byte, nano.BalanceSize)
	if _, err = reader.Read(balance); err != nil {
		return err
	}
	if err = b.Balance.UnmarshalBinary(balance); err != nil {
		return err
	}

	if _, err = reader.Read(b.Link[:]); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return b.Common.UnmarshalBinary(commonBytes)
}

func (b *StateBlock) Hash() Hash {
	return hashBytes(b.Address[:], b.PreviousHash[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:])
}

func (b *StateBlock) Root() Hash {
	if !b.PreviousHash.IsZero() {
		return b.PreviousHash
	}

	return b.Link
}

func (b *StateBlock) Signature() Signature {
	return b.Common.Signature
}

func (b *StateBlock) Size() int {
	return blockSizeState
}

func (b *StateBlock) ID() byte {
	return idBlockState
}

func (b *StateBlock) Valid() bool {
	if !b.PreviousHash.IsZero() {
		return b.Common.Work.Valid(b.PreviousHash)
	}

	return b.Common.Work.Valid(Hash(b.Address))
}
