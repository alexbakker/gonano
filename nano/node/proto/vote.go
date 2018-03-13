package proto

import "github.com/alexbakker/gonano/nano/block"

type VotePacket struct {
	Type byte
	Vote block.Vote
}
type ConfirmAckPacket VotePacket

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
