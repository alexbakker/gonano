package proto

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"net"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

const (
	idPacketInvalid byte = iota
	idPacketNotAType
	idPacketKeepAlive
	idPacketPublish
	idPacketConfirmReq
	idPacketConfirmAck
	idPacketBulkPull
	idPacketBulkPush
	idPacketFrontierReq
	idPacketBulkPullBlocks
)

type BulkPullMode byte

const (
	BulkPullModeList BulkPullMode = iota
	BulkPullModeChecksum
)

const (
	HeaderSize = 8

	VersionMax   = 0x06
	VersionUsing = 0x06
	VersionMin   = 0x04
)

var (
	ErrBadMagic  = errors.New("bad magic")
	ErrBadType   = errors.New("bad packet type")
	ErrBadLength = errors.New("bad packet length")

	Magic = [...]byte{'R', 'C'}

	packetNames = map[byte]string{
		idPacketInvalid:        "INVALID",
		idPacketNotAType:       "NOT_A_TYPE",
		idPacketKeepAlive:      "KEEP_ALIVE",
		idPacketPublish:        "PUBLISH",
		idPacketConfirmReq:     "CONFIRM_REQ",
		idPacketConfirmAck:     "CONFIRM_ACK",
		idPacketBulkPull:       "BULK_PULL",
		idPacketBulkPush:       "BULK_PUSH",
		idPacketFrontierReq:    "FRONTIER_REQ",
		idPacketBulkPullBlocks: "BULK_PULL_BLOCKS",
	}
)

type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	ID() byte
}

type Header struct {
	Magic        [2]byte
	VersionMax   byte
	VersionUsing byte
	VersionMin   byte
	MessageType  byte
	Extensions   uint16
}

type KeepAlivePacket struct {
	Peers []*net.UDPAddr
}

type FrontierReqPacket struct {
	StartAddress wallet.Address
	Age          uint32
	Count        uint32
}

type VotePacket struct {
	Type byte
	Vote block.Vote
}
type ConfirmAckPacket VotePacket

type BlockPacket struct {
	Type  byte
	Block block.Block
}
type ConfirmReqPacket BlockPacket
type PublishPacket BlockPacket

type BulkPullPacket struct {
	Address wallet.Address
	Hash    block.Hash
}

type BulkPullBlocksPacket struct {
	Min   block.Hash
	Max   block.Hash
	Mode  BulkPullMode
	Count uint32
}

func NewHeader(packetType byte) *Header {
	return &Header{
		Magic:        Magic,
		VersionMax:   VersionMax,
		VersionUsing: VersionUsing,
		VersionMin:   VersionMin,
		MessageType:  packetType,
	}
}

func NewKeepAlivePacket(peers []*net.UDPAddr) *KeepAlivePacket {
	return &KeepAlivePacket{
		Peers: peers,
	}
}

func Name(id byte) string {
	return packetNames[id]
}

func Parse(data []byte) (Packet, error) {
	if len(data) < HeaderSize {
		return nil, ErrBadLength
	}

	header := Header{}
	if err := header.UnmarshalBinary(data[:HeaderSize]); err != nil {
		return nil, err
	}

	// check the magic
	// todo: check version
	if !bytes.Equal(header.Magic[:], Magic[:]) {
		return nil, ErrBadMagic
	}

	// strip off the header
	data = data[HeaderSize:]

	var packet Packet
	switch header.MessageType {
	case idPacketKeepAlive:
		packet = new(KeepAlivePacket)
	case idPacketPublish:
		packet = &PublishPacket{Type: header.BlockType()}
	case idPacketConfirmReq:
		packet = &ConfirmReqPacket{Type: header.BlockType()}
	case idPacketConfirmAck:
		packet = &ConfirmAckPacket{Type: header.BlockType()}
	case idPacketBulkPull:
		packet = new(BulkPullPacket)
	//case idPacketBulkPush:
	case idPacketFrontierReq:
		packet = new(FrontierReqPacket)
	case idPacketBulkPullBlocks:
		packet = new(BulkPullBlocksPacket)
	default:
		return nil, ErrBadType
	}

	if err := packet.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return packet, nil
}

func MarshalPacket(packet Packet) ([]byte, error) {
	header := NewHeader(packet.ID())
	headerBytes, err := header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	packetBytes, err := packet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(headerBytes, packetBytes...), nil
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

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s BlockPacket) MarshalBinary() ([]byte, error) {
	return s.Block.MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *BlockPacket) UnmarshalBinary(data []byte) error {
	block, err := block.New(s.Type)
	if err != nil {
		return err
	}

	if err := block.UnmarshalBinary(data); err != nil {
		return err
	}

	s.Block = block
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s PublishPacket) MarshalBinary() ([]byte, error) {
	return BlockPacket(s).MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *PublishPacket) UnmarshalBinary(data []byte) error {
	packet := BlockPacket(*s)
	if err := packet.UnmarshalBinary(data); err != nil {
		return err
	}
	*s = PublishPacket(packet)
	return nil
}

func (s *PublishPacket) ID() byte {
	return idPacketPublish
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s ConfirmReqPacket) MarshalBinary() ([]byte, error) {
	return BlockPacket(s).MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *ConfirmReqPacket) UnmarshalBinary(data []byte) error {
	packet := BlockPacket(*s)
	if err := packet.UnmarshalBinary(data); err != nil {
		return err
	}
	*s = ConfirmReqPacket(packet)
	return nil
}

func (s *ConfirmReqPacket) ID() byte {
	return idPacketConfirmReq
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

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s VotePacket) MarshalBinary() ([]byte, error) {
	return s.Vote.MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *VotePacket) UnmarshalBinary(data []byte) error {
	block, err := block.New(s.Type)
	if err != nil {
		return err
	}

	s.Vote.Block = block
	return s.Vote.UnmarshalBinary(data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s ConfirmAckPacket) MarshalBinary() ([]byte, error) {
	return VotePacket(s).MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *ConfirmAckPacket) UnmarshalBinary(data []byte) error {
	packet := VotePacket(*s)
	if err := packet.UnmarshalBinary(data); err != nil {
		return err
	}
	*s = ConfirmAckPacket(packet)
	return nil
}

func (s *ConfirmAckPacket) ID() byte {
	return idPacketConfirmAck
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *BulkPullPacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(s.Address); err != nil {
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

	s.Address = make(wallet.Address, wallet.AddressSize)
	if _, err := reader.Read(s.Address); err != nil {
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
