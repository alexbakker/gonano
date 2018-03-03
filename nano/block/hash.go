package block

import (
	"bytes"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

const (
	HashSize = blake2b.Size256
)

type Hash [HashSize]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

func (h Hash) Equal(hash Hash) bool {
	return bytes.Equal(h[:], hash[:])
}
