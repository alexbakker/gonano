package proto

import (
	"fmt"
	"net"
)

type Network rune

const (
	NetworkTest = 'A'
	NetworkBeta = 'B'
	NetworkLive = 'C'
)

var (
	networkNames = map[Network]string{
		NetworkTest: "test",
		NetworkBeta: "beta",
		NetworkLive: "live",
	}
)

type Proto struct {
	net      Network
	magic    [2]byte
	versions Versions
}

type Versions struct {
	Max   byte
	Using byte
	Min   byte
}

func New(net Network) *Proto {
	return &Proto{
		net:   net,
		magic: [...]byte{'R', byte(net)},
		versions: Versions{
			Max:   0x07,
			Using: 0x07,
			Min:   0x07,
		},
	}
}

func (p *Proto) NewHeader(packetType byte) *Header {
	return &Header{
		Magic:        p.magic,
		VersionMax:   p.versions.Max,
		VersionUsing: p.versions.Using,
		VersionMin:   p.versions.Min,
		MessageType:  packetType,
	}
}

func (p *Proto) NewKeepAlivePacket(peers []*net.UDPAddr) *KeepAlivePacket {
	return &KeepAlivePacket{
		Peers: peers,
	}
}

func (p *Proto) UnmarshalPacket(data []byte) (Packet, error) {
	if len(data) < HeaderSize {
		return nil, ErrBadLength
	}

	header := Header{}
	if err := header.UnmarshalBinary(data[:HeaderSize]); err != nil {
		return nil, err
	}

	// check the magic
	// todo: check version
	if header.Magic != p.magic {
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
	case idPacketNodeIDHandshake:
		packet = new(HandshakePacket)
	default:
		return nil, ErrBadType
	}

	if err := packet.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return packet, nil
}

func (p *Proto) MarshalPacket(packet Packet) ([]byte, error) {
	header := p.NewHeader(packet.ID())

	switch t := packet.(type) {
	case *ConfirmReqPacket:
		header.SetBlockType(t.Type)
	case *PublishPacket:
		header.SetBlockType(t.Type)
	}

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

// MarshalText implements the encoding.TextMarshaler interface.
func (n Network) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (n *Network) UnmarshalText(text []byte) error {
	s := string(text)
	for k, v := range networkNames {
		if v == s {
			*n = k
			return nil
		}
	}

	return fmt.Errorf("unknown network: %s", s)
}

// String implements the fmt.Stringer interface.
func (n Network) String() string {
	s, ok := networkNames[n]
	if !ok {
		return "unknown"
	}

	return s
}
