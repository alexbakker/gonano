package proto

import (
	"net"
)

type Network rune

const (
	NetworkTest = 'A'
	NetworkBeta = 'B'
	NetworkLive = 'C'
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
			Max:   0x06,
			Using: 0x06,
			Min:   0x04,
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
