package proto

import "github.com/alexbakker/gonano/nano/block"

type BlockPacket struct {
	Type  byte
	Block block.Block
}
type ConfirmReqPacket BlockPacket
type PublishPacket BlockPacket

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
