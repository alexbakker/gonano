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
	// Flush flushes all items in the cache through the registered callback.
	Flush()
	// ReadNext reads the next packet from the given reader, parses it and adds
	// the item to the cache.
	ReadNext(r io.Reader) (done bool, err error)
	// WriteNext writes the next packet to the given reader. If no packets need
	// to be sent, 'done' is set to true.
	WriteNext(w io.Writer) (done bool, err error)
}

type (
	FrontierSyncerFunc       func(*block.Frontier)
	BulkPullSyncerFunc       func([]block.Block)
	BulkPullBlocksSyncerFunc func([]block.Block)
)

type FrontierSyncer struct {
	cb FrontierSyncerFunc
}

type BulkPullSyncer struct {
	blocks     []block.Block
	frontiers  []*block.Frontier
	writeIndex int
	readIndex  int
	cb         BulkPullSyncerFunc
}

type BulkPullBlocksSyncer struct {
	blocks []block.Block
	mode   proto.BulkPullMode
	cb     BulkPullBlocksSyncerFunc
}

func NewFrontierSyncer(cb FrontierSyncerFunc) *FrontierSyncer {
	return &FrontierSyncer{cb: cb}
}

func NewBulkPullSyncer(cb BulkPullSyncerFunc, frontiers []*block.Frontier) *BulkPullSyncer {
	return &BulkPullSyncer{
		cb:        cb,
		frontiers: frontiers,
		blocks:    make([]block.Block, 0, syncCacheSize),
	}
}

func NewBulkPullBlocksSyncer(cb BulkPullBlocksSyncerFunc) *BulkPullBlocksSyncer {
	return &BulkPullBlocksSyncer{
		cb:     cb,
		mode:   proto.BulkPullModeList,
		blocks: make([]block.Block, 0, syncCacheSize),
	}
}

func Sync(syncer Syncer, peer *Peer) error {
	conn, err := initSync(peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	errs := make(chan error)
	cancel := make(chan struct{})

	go func(errs chan<- error, cancel <-chan struct{}) {
		for {
			select {
			case <-cancel:
				errs <- nil
				return
			default:
				// continue
			}

			// reset the write deadline
			if err := conn.SetWriteDeadline(time.Now().Add(syncTimeout)); err != nil {
				errs <- err
				return
			}

			// allow the syncer to send some data if it needs to
			done, err := syncer.WriteNext(conn)
			if err != nil {
				errs <- err
				return
			}

			if done {
				errs <- nil
				return
			}
		}
	}(errs, cancel)

	go func(errs chan<- error, cancel <-chan struct{}) {
		reader := bufio.NewReader(conn)

		for {
			select {
			case <-cancel:
				errs <- nil
				return
			default:
				// continue
			}

			// reset the read deadline
			if err := conn.SetReadDeadline(time.Now().Add(syncTimeout)); err != nil {
				errs <- err
				return
			}

			// let the syncer read the next packet
			done, err := syncer.ReadNext(reader)
			if err != nil {
				errs <- err
				return
			}

			if done {
				errs <- nil
				return
			}
		}
	}(errs, cancel)

	// wait for both goroutines to finish or for one of them to error out
	var res error
	for i := 0; i < 2; {
		select {
		case err := <-errs:
			// if an error occurred, cancel the other goroutine
			if err != nil {
				res = err
				close(cancel)
			}
			i++
		}
	}
	if res != nil {
		return err
	}

	// flush the syncer cache
	syncer.Flush()
	return nil
}

// ReadNext implements the Syncer interface.
func (s *FrontierSyncer) ReadNext(r io.Reader) (bool, error) {
	var buf [block.FrontierSize]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return false, err
	}

	var frontier block.Frontier
	if err := frontier.UnmarshalBinary(buf[:]); err != nil {
		return false, err
	}

	// if the frontier is zero, the transmission is complete
	if frontier.IsZero() {
		return true, nil
	}

	// report to the caller
	s.cb(&frontier)

	return false, nil
}

// WriteNext implements the Syncer interface.
func (s *FrontierSyncer) WriteNext(w io.Writer) (done bool, err error) {
	packet := proto.FrontierReqPacket{
		StartAddress: make(wallet.Address, wallet.AddressSize),
		Age:          math.MaxUint32,
		Count:        math.MaxUint32,
	}

	return true, writePacket(w, &packet)
}

// Flush implements the Syncer interface.
func (s *FrontierSyncer) Flush() {

}

// ReadNext implements the Syncer interface.
func (s *BulkPullSyncer) ReadNext(r io.Reader) (bool, error) {
	blk, err := readBlock(r)
	if err != nil {
		if err == block.ErrNotABlock {
			s.readIndex++
			if s.readIndex >= len(s.frontiers) {
				return true, nil
			}
			return false, nil
		}
		return false, err
	}

	// report to the caller if the cache is full
	s.blocks = append(s.blocks, blk)
	if len(s.blocks) >= syncCacheSize {
		s.Flush()
	}

	return false, nil
}

// WriteNext implements the Syncer interface.
func (s *BulkPullSyncer) WriteNext(w io.Writer) (done bool, err error) {
	if s.writeIndex < len(s.frontiers) {
		// request the chain of the next frontier
		packet := proto.BulkPullPacket{
			Address: s.frontiers[s.writeIndex].Address,
		}

		s.writeIndex++
		return false, writePacket(w, &packet)
	}

	return true, nil
}

// Flush implements the Syncer interface.
func (s *BulkPullSyncer) Flush() {
	if len(s.blocks) > 0 {
		s.cb(s.blocks)
		s.blocks = s.blocks[:0]
	}
}

// ReadNext implements the Syncer interface.
func (s *BulkPullBlocksSyncer) ReadNext(r io.Reader) (bool, error) {
	blk, err := readBlock(r)
	if err != nil {
		if err == block.ErrNotABlock {
			return true, nil
		}
		return false, err
	}

	// report to the caller if the cache is full
	s.blocks = append(s.blocks, blk)
	if len(s.blocks) >= syncCacheSize {
		s.Flush()
	}

	return false, nil
}

// WriteNext implements the Syncer interface.
func (s *BulkPullBlocksSyncer) WriteNext(w io.Writer) (done bool, err error) {
	packet := proto.BulkPullBlocksPacket{
		Mode:  s.mode,
		Count: math.MaxUint32,
	}

	// set min hash to zero and the max hash to math.MaxUint8
	for i := 0; i < block.HashSize; i++ {
		packet.Min[i] = 0
		packet.Max[i] = math.MaxUint8
	}

	return true, writePacket(w, &packet)
}

// Flush implements the Syncer interface.
func (s *BulkPullBlocksSyncer) Flush() {
	if len(s.blocks) > 0 {
		s.cb(s.blocks)
		s.blocks = s.blocks[:0]
	}
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

func writePacket(w io.Writer, packet proto.Packet) error {
	// todo: get rid of this proto.New mess
	packetBytes, err := proto.New(proto.NetworkLive).MarshalPacket(packet)
	if err != nil {
		return err
	}

	_, err = w.Write(packetBytes)
	return err
}

func readBlock(r io.Reader) (block.Block, error) {
	var head [1]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return nil, err
	}

	blk, err := block.New(head[0])
	if err != nil {
		return nil, err
	}

	buf := make([]byte, blk.Size())
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	if err := blk.UnmarshalBinary(buf); err != nil {
		return nil, err
	}

	// skip blocks with invalid work
	// todo: properly handle invalid blocks
	if !blk.Valid() {
		fmt.Printf("bad work for block: %s\n", blk.Hash())
		return nil, err
	}

	return blk, err
}
