package node

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/node/proto"
	"github.com/alexbakker/gonano/nano/store"
)

var (
	errBadIP        = errors.New("bad ip")
	errIPv6Disabled = errors.New("tried to use ipv6 while it's disabled")
	errBadProtocol  = errors.New("unexpected protocol for this packet")

	DefaultOptions = Options{
		Address:      ":7075",
		EnableIPv6:   false,
		EnableVoting: true,
		MaxPeers:     15,
	}
)

type Node struct {
	options Options
	udpConn *net.UDPConn
	tcpConn *net.TCPListener
	peers   *PeerList
	ledger  *store.Ledger

	frontiers []*block.Frontier
}

type Options struct {
	Address      string
	EnableIPv6   bool
	EnableVoting bool
	MaxPeers     int
	Peers        []*net.UDPAddr
}

func New(ledger *store.Ledger, options Options) (*Node, error) {
	// setup the udp listener
	udpAddr, err := net.ResolveUDPAddr("udp", options.Address)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	// setup the tcp listener
	tcpAddr, err := net.ResolveTCPAddr("tcp", options.Address)
	if err != nil {
		return nil, err
	}
	tcpConn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	return &Node{
		udpConn: udpConn,
		tcpConn: tcpConn,
		options: options,
		peers:   NewPeerList(options.MaxPeers),
		ledger:  ledger,
	}, nil
}

func (n *Node) Run() error {
	for _, addr := range n.options.Peers {
		if _, err := n.addPeer(addr); err != nil {
			return err
		}
	}

	go n.syncFontiers()
	go n.syncBlocks()

	return n.listenUDP()
}

func (n *Node) Stop() error {
	// stop listening
	var err error
	if err = n.udpConn.Close(); err != nil {
		return err
	}
	if err = n.tcpConn.Close(); err != nil {
		return err
	}

	return nil
}

func (n *Node) listenUDP() error {
	buf := make([]byte, 1024)
	for {
		recv, addr, err := n.udpConn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		data := buf[:recv]
		packet, err := proto.Parse(data)
		if err != nil {
			fmt.Printf("recv error: %s\n", err)
			continue
		}

		// todo: remove
		fmt.Printf("recv packet (%s): %s (%d bytes)\n", addr.String(), proto.Name(packet.ID()), len(data))

		if err := n.handlePacket(addr, packet); err != nil {
			fmt.Printf("error handling packet: %s\n", err)
			continue
		}
	}

	return nil
}

func (n *Node) listenTCP() error {
	return errors.New("not implemented")
}

// syncFrontiers asks a random peer for a list of frontiers once every 5
// minutes.
func (n *Node) syncFontiers() error {
	var startTime time.Time

	for {
		startTime = time.Now()

		// pick a random peer
		peer, err := n.peers.Random()
		if err != nil {
			// todo: handle this better
			fmt.Printf("error picking random peer: %s\n", err)
			break
		}

		fmt.Printf("requesting frontiers from: %s\n", peer.Addr)

		syncer := NewFrontierSyncer(n.processFrontier)
		if err = Sync(syncer, peer); err == nil {
			fmt.Printf("received frontiers: %d\n", len(n.frontiers))
			syncer := NewBulkPullSyncer(n.processFrontierBlocks, n.frontiers)
			if err := Sync(syncer, peer); err == nil {
				if count, err := n.ledger.CountBlocks(); err == nil {
					fmt.Printf("block count: %d\n", count)
				}
			}
		}

		// retry sooner if an error occurred
		if err == nil {
			delta := time.Minute*5 - time.Since(startTime)
			if delta > 0 {
				time.Sleep(delta)
			}
		} else {
			fmt.Printf("error requesting frontiers: %s\n", err)
			time.Sleep(time.Second * 2)
		}
	}

	return nil
}

func (n *Node) syncBlocks() error {
	return nil
}

func (n *Node) processFrontier(frontier *block.Frontier) {
	fmt.Printf("frontier: %+v\n", frontier)
	n.frontiers = append(n.frontiers, frontier)
}

func (n *Node) processFrontierBlocks(blocks []block.Block) {
	// if we feed the list of blocks to the ledger in reverse, there's a good
	// chance the blocks are magically in the right order

	// note: this modifies the original slice
	for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
		blocks[i], blocks[j] = blocks[j], blocks[i]
	}

	if err := n.ledger.AddBlocks(blocks); err != nil {
		//fmt.Printf("error adding block: %s\n", err)
	}
}

func (n *Node) processBlocks(blocks []block.Block) {
	/*if err := n.ledger.AddBlocks(blocks); err != nil { fmt.Printf("error
		processing blocks: %s\n", err)
		return
	}*/

	for _, blk := range blocks {
		fmt.Printf("block (%s) %s: %+v\n", block.Name(blk.ID()), blk.Hash(), blk)
	}
}

func (n *Node) addPeer(addr *net.UDPAddr) (*Peer, error) {
	if !addr.IP.IsGlobalUnicast() {
		return nil, errBadIP
	}

	// don't add ipv6 peers if ipv6 is disabled
	if addr.IP.To4() == nil && !n.options.EnableIPv6 {
		return nil, errIPv6Disabled
	}

	peer, err := n.peers.Add(addr)
	if err != nil {
		return nil, err
	}

	// if sending a keep alive packet fails, remove it from the list again
	if err := n.sendKeepAlive(peer); err != nil {
		n.peers.Remove(peer)
		return nil, err
	}

	fmt.Printf("add peer: %s\n", peer.Addr)
	return peer, nil
}

func (n *Node) sendPacket(addr *net.UDPAddr, packet proto.Packet) error {
	bytes, err := proto.MarshalPacket(packet)
	if err != nil {
		return err
	}

	_, err = n.udpConn.WriteToUDP(bytes, addr)

	// todo: remove
	fmt.Printf("send (%s): %s (%d bytes)\n", addr.String(), proto.Name(packet.ID()), len(bytes))

	return err
}

func (n *Node) sendKeepAlive(target *Peer) error {
	// pick a couple of random peers to share
	peers, err := n.peers.Pick()
	if err != nil {
		return err
	}

	var addrs []*net.UDPAddr
	for _, p := range peers {
		// don't add the target peer to the list
		if p == target {
			continue
		}
		addrs = append(addrs, p.Addr)
	}

	packet := proto.NewKeepAlivePacket(addrs)
	return n.sendPacket(target.Addr, packet)
}

func (n *Node) handlePacket(addr *net.UDPAddr, packet proto.Packet) error {
	switch p := packet.(type) {
	case *proto.KeepAlivePacket:
		return n.handleKeepAlivePacket(addr, p)
	case *proto.ConfirmAckPacket:
	case *proto.ConfirmReqPacket:
	case *proto.PublishPacket:
	default:
		return errBadProtocol
	}
	return nil
}

func (n *Node) handleKeepAlivePacket(addr *net.UDPAddr, packet *proto.KeepAlivePacket) error {
	peer := n.peers.Get(addr)
	if peer != nil {
		// if we know about this peer, send a keep alive packet back if it's been a while
		err := peer.Ping(func() error {
			return n.sendKeepAlive(peer)
		})
		if err != nil {
			return err
		}
	} else if !n.peers.Full() {
		// if we don't know about this peer, try adding it to our list
		if _, err := n.addPeer(addr); err != nil {
			return err
		}
	}

	// add any peers we don't already know about to our list
	for _, peerAddr := range packet.Peers {
		if n.peers.Full() || n.peers.Get(peerAddr) != nil {
			break
		}

		if _, err := n.addPeer(peerAddr); err != nil {
			return err
		}
	}

	return nil
}
