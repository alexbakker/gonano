package block

import (
	"encoding/hex"

	"github.com/alexbakker/gonano/nano/crypto/ed25519"
)

const (
	SignatureSize = ed25519.SignatureSize
)

type Signature [SignatureSize]byte

func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}
