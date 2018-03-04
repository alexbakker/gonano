package node

import (
	"net"
	"time"
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

// Ping will call the given function if the peer needs to be pinged. If fn
// returns nil, the last ping time is reset.
func (p *Peer) Ping(fn func() error) error {
	if time.Since(p.lastPing) > peerTimeout/2 {
		if err := fn(); err != nil {
			return err
		}

		p.lastPing = time.Now()
	}

	return nil
}

// Stale reports whether it's been a while since we've received a keep alive
// packet from this peer.
func (p *Peer) Stale() bool {
	return time.Since(p.lastPong) > peerTimeout/2
}

// Dead reports whether this peer should be considered dead and be removed from
// the peer list.
func (p *Peer) Dead() bool {
	return time.Since(p.lastPong) > peerTimeout
}

// Pong resets the pong timeout for this peer. It should be called when we've
// received a keep alive packet from this peer.
func (p *Peer) Pong() {
	p.lastPong = time.Now()
}
