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
	preambleSize     = 32
	blockSizeCommon  = SignatureSize + WorkSize
	blockSizeOpen    = blockSizeCommon + HashSize + nano.AddressSize*2
	blockSizeSend    = blockSizeCommon + HashSize + nano.AddressSize + nano.BalanceSize
	blockSizeReceive = blockSizeCommon + HashSize*2
	blockSizeChange  = blockSizeCommon + HashSize + nano.AddressSize
	blockSizeState   = blockSizeCommon + HashSize*2 + nano.AddressSize*2 + nano.BalanceSize
)

type Block interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Hash() Hash
	Root() Hash
	Size() int
	ID() byte
	Valid(threshold uint64) bool
}

type OpenBlock struct {
	SourceHash     Hash         `json:"source"`
	Representative nano.Address `json:"representative"`
	Address        nano.Address `json:"address"`
	Signature      Signature    `json:"signature"`
	Work           Work         `json:"work"`
}

type SendBlock struct {
	PreviousHash Hash         `json:"previous"`
	Destination  nano.Address `json:"destination"`
	Balance      nano.Balance `json:"balance"`
	Signature    Signature    `json:"signature"`
	Work         Work         `json:"work"`
}

type ReceiveBlock struct {
	PreviousHash Hash      `json:"previous"`
	SourceHash   Hash      `json:"source"`
	Signature    Signature `json:"signature"`
	Work         Work      `json:"work"`
}

type ChangeBlock struct {
	PreviousHash   Hash         `json:"previous"`
	Representative nano.Address `json:"representative"`
	Signature      Signature    `json:"signature"`
	Work           Work         `json:"work"`
}

type StateBlock struct {
	Address        nano.Address `json:"address"`
	PreviousHash   Hash         `json:"previous"`
	Representative nano.Address `json:"representative"`
	Balance        nano.Balance `json:"balance"`
	Link           Hash         `json:"link"`
	Signature      Signature    `json:"signature"`
	Work           Work         `json:"work"`
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
	case idBlockState:
		return new(StateBlock), nil
	case idBlockNotABlock:
		return nil, ErrNotABlock
	default:
		return nil, ErrBadBlockType
	}
}

func Name(id byte) string {
	return blockNames[id]
}

func marshalCommon(sig Signature, work Work, order binary.ByteOrder) ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(sig[:]); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, order, work); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func unmarshalCommon(data []byte, order binary.ByteOrder, sig *Signature, work *Work) error {
	reader := bytes.NewReader(data)

	var s Signature
	if _, err := reader.Read(s[:]); err != nil {
		return err
	}
	*sig = s

	var w Work
	if err := binary.Read(reader, order, &w); err != nil {
		return err
	}
	*work = w

	if err := util.AssertReaderEOF(reader); err != nil {
		return err
	}

	return nil
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

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.LittleEndian)
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

	return unmarshalCommon(commonBytes, binary.LittleEndian, &b.Signature, &b.Work)
}

func (b *OpenBlock) Hash() Hash {
	return hashBytes(b.SourceHash[:], b.Representative[:], b.Address[:])
}

func (b *OpenBlock) Root() Hash {
	return b.SourceHash
}

func (b *OpenBlock) Size() int {
	return blockSizeOpen
}

func (b *OpenBlock) ID() byte {
	return idBlockOpen
}

func (b *OpenBlock) Valid(threshold uint64) bool {
	return b.Work.Valid(Hash(b.Address), threshold)
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

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.LittleEndian)
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

	return unmarshalCommon(commonBytes, binary.LittleEndian, &b.Signature, &b.Work)
}

func (b *SendBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.Destination[:], b.Balance.Bytes(binary.BigEndian))
}

func (b *SendBlock) Root() Hash {
	return b.PreviousHash
}

func (b *SendBlock) Size() int {
	return blockSizeSend
}

func (b *SendBlock) ID() byte {
	return idBlockSend
}

func (b *SendBlock) Valid(threshold uint64) bool {
	return b.Work.Valid(b.PreviousHash, threshold)
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

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.LittleEndian)
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

	return unmarshalCommon(commonBytes, binary.LittleEndian, &b.Signature, &b.Work)
}

func (b *ReceiveBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.SourceHash[:])
}

func (b *ReceiveBlock) Root() Hash {
	return b.PreviousHash
}

func (b *ReceiveBlock) Size() int {
	return blockSizeReceive
}

func (b *ReceiveBlock) ID() byte {
	return idBlockReceive
}

func (b *ReceiveBlock) Valid(threshold uint64) bool {
	return b.Work.Valid(b.PreviousHash, threshold)
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

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.LittleEndian)
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

	return unmarshalCommon(commonBytes, binary.LittleEndian, &b.Signature, &b.Work)
}

func (b *ChangeBlock) Hash() Hash {
	return hashBytes(b.PreviousHash[:], b.Representative[:])
}

func (b *ChangeBlock) Root() Hash {
	return b.PreviousHash
}

func (b *ChangeBlock) Size() int {
	return blockSizeChange
}

func (b *ChangeBlock) ID() byte {
	return idBlockChange
}

func (b *ChangeBlock) Valid(threshold uint64) bool {
	return b.Work.Valid(b.PreviousHash, threshold)
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

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.BigEndian)
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

	return unmarshalCommon(commonBytes, binary.BigEndian, &b.Signature, &b.Work)
}

func (b *StateBlock) Hash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = idBlockState
	return hashBytes(preamble[:], b.Address[:], b.PreviousHash[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:])
}

func (b *StateBlock) Root() Hash {
	if !b.IsOpen() {
		return b.PreviousHash
	}

	return b.Link
}

func (b *StateBlock) Size() int {
	return blockSizeState
}

func (b *StateBlock) ID() byte {
	return idBlockState
}

func (b *StateBlock) Valid(threshold uint64) bool {
	if !b.IsOpen() {
		return b.Work.Valid(b.PreviousHash, threshold)
	}

	return b.Work.Valid(Hash(b.Address), threshold)
}

func (b *StateBlock) IsOpen() bool {
	return b.PreviousHash.IsZero()
}
