package node

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/node/proto"
	"github.com/alexbakker/gonano/nano/wallet"
)

const (
	syncTimeout   = time.Second * 2
	syncCacheSize = 10000
)

type Syncer interface {
	// Parse parses the given packet.
	Parse(buf []byte) (bool, error)
	// Size returns the size of the next packet, given the packet header. If it
	// returns a size of 0 and no error, the transmission is complete.
	Size(head []byte) (int, error)
	// HeadSize returns the size that the header for each packet will have.
	HeadSize() int
	// NextPacket returns the next packet that should be sent to continue the
	// sync. If it returns nil, the sync is complete.
	NextPacket() proto.Packet
	// Flush flushes all items in the cache through the registered callback.
	Flush()
}

type (
	FrontierSyncerFunc       func(*block.Frontier)
	BulkPullSyncerFunc       func([]block.Block)
	BulkPullBlocksSyncerFunc func([]block.Block)
)

type FrontierSyncer struct {
	current *block.Frontier
	cb      FrontierSyncerFunc
	sent    bool
}

type BulkPullSyncer struct {
	blocks    []block.Block
	current   block.Block
	frontiers []*block.Frontier
	i         int
	cb        BulkPullSyncerFunc
}

type BulkPullBlocksSyncer struct {
	blocks  []block.Block
	current block.Block
	sent    bool
	mode    proto.BulkPullMode
	cb      BulkPullBlocksSyncerFunc
}

func NewFrontierSyncer(cb FrontierSyncerFunc) *FrontierSyncer {
	return &FrontierSyncer{cb: cb}
}

func NewBulkPullSyncer(cb BulkPullSyncerFunc, frontiers []*block.Frontier) *BulkPullSyncer {
	return &BulkPullSyncer{cb: cb, frontiers: frontiers}
}

func NewBulkPullBlocksSyncer(cb BulkPullBlocksSyncerFunc) *BulkPullBlocksSyncer {
	return &BulkPullBlocksSyncer{cb: cb, mode: proto.BulkPullModeList}
}

func Sync(syncer Syncer, peer *Peer) error {
	conn, err := initSync(peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	packet := syncer.NextPacket()
	if err := sendPacket(conn, packet); err != nil {
		return err
	}

	buf := make([]byte, 256)
	headSize := syncer.HeadSize()
	head := make([]byte, headSize)
	reader := bufio.NewReader(conn)

	for {
		if headSize > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(syncTimeout)); err != nil {
				return err
			}
			if _, err := io.ReadFull(reader, head); err != nil {
				return err
			}
		}

		size, err := syncer.Size(head)
		if err != nil {
			return err
		}

		var isDone bool
		if size > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(syncTimeout)); err != nil {
				return err
			}
			if _, err := io.ReadFull(reader, buf[:size]); err != nil {
				return err
			}

			done, err := syncer.Parse(buf[:size])
			if err != nil {
				return err
			}

			isDone = done
		}

		packet := syncer.NextPacket()
		if packet != nil {
			if err := sendPacket(conn, packet); err != nil {
				return err
			}
		} else if isDone {
			syncer.Flush()
			return nil
		}
	}
}

// Parse implements the Syncer interface.
func (s *FrontierSyncer) Parse(buf []byte) (bool, error) {
	s.current = new(block.Frontier)
	if err := s.current.UnmarshalBinary(buf); err != nil {
		return false, err
	}

	// if the frontier is zero, the transmission is complete
	if s.current.IsZero() {
		return true, nil
	}

	// report to the caller
	s.cb(s.current)

	return false, nil
}

// Flush implements the Syncer interface.
func (s *FrontierSyncer) Flush() {

}

// Size implements the Syncer interface.
func (s *FrontierSyncer) Size(head []byte) (int, error) {
	return block.FrontierSize, nil
}

// HeadSize implements the Syncer interface.
func (s *FrontierSyncer) HeadSize() int {
	return 0
}

// NextPacket implements the Syncer interface.
func (s *FrontierSyncer) NextPacket() proto.Packet {
	if !s.sent {
		packet := &proto.FrontierReqPacket{
			StartAddress: make(wallet.Address, wallet.AddressSize),
			Age:          math.MaxUint32,
			Count:        math.MaxUint32,
		}

		s.sent = true
		return packet
	}

	return nil
}

// Parse implements the Syncer interface.
func (s *BulkPullSyncer) Parse(buf []byte) (bool, error) {
	if err := s.current.UnmarshalBinary(buf); err != nil {
		return false, err
	}

	// skip blocks with invalid work
	// todo: properly handle invalid blocks
	if !s.current.Valid() {
		fmt.Printf("bad work for block: %s\n", s.current.Hash())
		return false, nil
	}

	s.blocks = append(s.blocks, s.current)
	return false, nil
}

// Flush implements the Syncer interface.
func (s *BulkPullSyncer) Flush() {
	if len(s.blocks) > 0 {
		s.cb(s.blocks)
		s.blocks = nil
	}
}

// Size implements the Syncer interface.
func (s *BulkPullSyncer) Size(head []byte) (int, error) {
	blk, err := block.New(head[0])
	if err != nil {
		// this indicates that the transmission is complete
		if err == block.ErrNotABlock {
			// report this frontier block list to the caller
			s.Flush()
			return 0, nil
		}
		return 0, err
	}

	s.current = blk
	return s.current.Size(), nil
}

// HeadSize implements the Syncer interface.
func (s *BulkPullSyncer) HeadSize() int {
	return 1
}

// NextPacket implements the Syncer interface.
func (s *BulkPullSyncer) NextPacket() proto.Packet {
	if s.i < len(s.frontiers) {
		// request the chain of the next frontier
		packet := &proto.BulkPullPacket{
			Address: s.frontiers[s.i].Address,
		}

		s.i++
		return packet
	}

	return nil
}

// Parse implements the Syncer interface.
func (s *BulkPullBlocksSyncer) Parse(buf []byte) (bool, error) {
	if err := s.current.UnmarshalBinary(buf); err != nil {
		return false, err
	}

	// skip blocks with invalid work
	// todo: properly handle invalid blocks
	if !s.current.Valid() {
		fmt.Printf("bad work for block: %s\n", s.current.Hash())
		return false, nil
	}

	// report to the caller if the cache if full
	s.blocks = append(s.blocks, s.current)
	if len(s.blocks) >= syncCacheSize {
		s.Flush()
	}

	return false, nil
}

// Flush implements the Syncer interface.
func (s *BulkPullBlocksSyncer) Flush() {
	if len(s.blocks) > 0 {
		s.cb(s.blocks)
		s.blocks = nil
	}
}

// Size implements the Syncer interface.
func (s *BulkPullBlocksSyncer) Size(head []byte) (int, error) {
	blk, err := block.New(head[0])
	if err != nil {
		// this indicates that the transmission is complete
		if err == block.ErrNotABlock {
			return 0, nil
		}
		return 0, err
	}

	s.current = blk
	return s.current.Size(), nil
}

// HeadSize implements the Syncer interface.
func (s *BulkPullBlocksSyncer) HeadSize() int {
	return 1
}

// NextPacket implements the Syncer interface.
func (s *BulkPullBlocksSyncer) NextPacket() proto.Packet {
	if !s.sent {
		packet := &proto.BulkPullBlocksPacket{
			Mode:  s.mode,
			Count: math.MaxUint32,
		}

		// set min hash to zero and the max hash to math.MaxUint8
		for i := 0; i < block.HashSize; i++ {
			packet.Min[i] = 0
			packet.Max[i] = math.MaxUint8
		}

		s.sent = true
		return packet
	}

	return nil
}

func initSync(peer *Peer) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", peer.Addr.String())
	if err != nil {
		return nil, err
	}

	conn, err := util.DialTCP(addr, syncTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendPacket(conn *net.TCPConn, packet proto.Packet) error {
	// todo: get rid of this proto.New mess
	packetBytes, err := proto.New(proto.NetworkLive).MarshalPacket(packet)
	if err != nil {
		return err
	}

	if err := conn.SetWriteDeadline(time.Now().Add(syncTimeout)); err != nil {
		return err
	}

	_, err = conn.Write(packetBytes)
	return err
}
