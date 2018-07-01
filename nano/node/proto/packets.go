package proto

import (
	"encoding"
	"errors"
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
	idPacketNodeIDHandshake
)

var (
	ErrBadMagic  = errors.New("bad magic")
	ErrBadType   = errors.New("bad packet type")
	ErrBadLength = errors.New("bad packet length")

	packetNames = map[byte]string{
		idPacketInvalid:         "invalid",
		idPacketNotAType:        "not_a_type",
		idPacketKeepAlive:       "keep_alive",
		idPacketPublish:         "publish",
		idPacketConfirmReq:      "confirm_req",
		idPacketConfirmAck:      "confirm_ack",
		idPacketBulkPull:        "bulk_pull",
		idPacketBulkPush:        "bulk_push",
		idPacketFrontierReq:     "frontier_req",
		idPacketBulkPullBlocks:  "bulk_pull_blocks",
		idPacketNodeIDHandshake: "node_id_handshake",
	}
)

type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	ID() byte
}

func Name(id byte) string {
	return packetNames[id]
}
