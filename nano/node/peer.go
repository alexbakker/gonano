package node

import (
	"errors"
	"net"
	"time"

	"github.com/alexbakker/gonano/nano/crypto/random"
)

var (
	ErrMaxPeers   = errors.New("max amount of peers reached")
	ErrPeerExists = errors.New("this peer already exists in the list")
)

const (
	peerTimeout = time.Second * 30
)

// Peer represents a Nano peer.
type Peer struct {
	Addr     *net.UDPAddr
	lastPing time.Time
	lastPong time.Time
}

// PeerList represents a list of peers.
type PeerList struct {
	peers []*Peer
	max   int
}

// NewPeerList creates a new peer list with the given maximum capacity.
func NewPeerList(max int) *PeerList {
	return &PeerList{max: max}
}

// Add creates a new peer instance with the given address, adds it to the
// internal peer list and returns it.
func (l *PeerList) Add(addr *net.UDPAddr) (*Peer, error) {
	// enforce a maximum amount of peers
	if l.Full() {
		return nil, ErrMaxPeers
	}

	// check if we already have this peer in our list
	if l.Get(addr) != nil {
		return nil, ErrPeerExists
	}

	peer := &Peer{Addr: addr}
	l.peers = append(l.peers, peer)
	return peer, nil
}

// Get retrieves a peer with the given address. If no such peer exists, nil is
// returned.
func (l *PeerList) Get(addr *net.UDPAddr) *Peer {
	for _, peer := range l.peers {
		if peer.Addr.IP.Equal(addr.IP) && peer.Addr.Port == addr.Port {
			return peer
		}
	}

	return nil
}

// Remove removes the given peer from the list.
func (l *PeerList) Remove(peer *Peer) {
	for i := range l.peers {
		if l.peers[i] == peer {
			l.peers = append(l.peers[:i], l.peers[i+1:]...)
			return
		}
	}
	panic("peer not in list")
}

// Full reports whether the internal peer list has reached its maximum capacity.
func (l *PeerList) Full() bool {
	return len(l.peers) >= l.max
}

// Len returns the length of the internal peer list.
func (l *PeerList) Len() int {
	return len(l.peers)
}

// Pick returns 8 random peers from the internal peer list. This function is
// usually used to populate a KeepAlivePacket.
func (l *PeerList) Pick() ([]*Peer, error) {
	var size int
	if len(l.peers) > 8 {
		size = 8
	} else {
		size = len(l.peers)
	}

	perm, err := random.Perm(len(l.peers))
	if err != nil {
		return nil, err
	}

	peers := make([]*Peer, size)
	for i := 0; i < size; i++ {
		peers[i] = l.peers[perm[i]]
	}

	return peers, nil
}

// Random picks one random peer from the internal peer list and returns it.
func (l *PeerList) Random() (*Peer, error) {
	i, err := random.Intn(len(l.peers))
	if err != nil {
		return nil, err
	}

	return l.peers[i], nil
}

// Peers returns a copy of the internal peer list.
func (l *PeerList) Peers() []*Peer {
	peers := make([]*Peer, len(l.peers))
	copy(peers, l.peers)
	return peers
}
