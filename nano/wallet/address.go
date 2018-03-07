package wallet

import (
	"bytes"
	"encoding/base32"
	"errors"
	"strings"

	"github.com/alexbakker/gonano/nano/crypto/ed25519"
	"github.com/alexbakker/gonano/nano/internal/util"
	"golang.org/x/crypto/blake2b"
)

const (
	// AddressLen represents the string length of a Nano address.
	AddressLen = 64
	// AddressSize represents the binary size of a Nano address (a public key).
	AddressSize = ed25519.PublicKeySize
	// AddressPrefix is the prefix of Nano addresses.
	AddressPrefix = "xrb_"

	// AddressEncodingAlphabet is Nano's custom alphabet for base32 encoding
	AddressEncodingAlphabet = "13456789abcdefghijkmnopqrstuwxyz"
)

var (
	// AddressEncoding is a base32 encoding using NanoEncodingAlphabet as its
	// alphabet.
	AddressEncoding = base32.NewEncoding(AddressEncodingAlphabet)

	ErrAddressLen      = errors.New("bad address length")
	ErrAddressPrefix   = errors.New("bad address prefix")
	ErrAddressEncoding = errors.New("bad address encoding")
	ErrAddressChecksum = errors.New("bad address checksum")
)

// Address represents a Nano address.
type Address ed25519.PublicKey

// ParseAddress parses the given Nano address string to a public key.
func ParseAddress(s string) (Address, error) {
	if len(s) != AddressLen {
		return nil, ErrAddressLen
	}

	if !strings.HasPrefix(s, AddressPrefix) {
		return nil, ErrAddressPrefix
	}

	key, err := AddressEncoding.DecodeString("1111" + s[4:56])
	if err != nil {
		return nil, ErrAddressEncoding
	}

	checksum, err := AddressEncoding.DecodeString(s[56:])
	if err != nil {
		return nil, ErrAddressEncoding
	}

	address := Address(ed25519.PublicKey(key[3:]))
	if !bytes.Equal(address.Checksum(), checksum) {
		return nil, ErrAddressChecksum
	}

	return address, nil
}

// Checksum calculates the checksum for this address' public key.
func (a Address) Checksum() []byte {
	hash, err := blake2b.New(5, nil)
	if err != nil {
		panic(err)
	}

	hash.Write(a)
	return util.ReverseBytes(hash.Sum(nil))
}

// String implements the fmt.Stringer interface.
func (a Address) String() string {
	key := append([]byte{0, 0, 0}, a...)
	encodedKey := AddressEncoding.EncodeToString(key)[4:]
	encodedChecksum := AddressEncoding.EncodeToString(a.Checksum())

	var buf bytes.Buffer
	buf.WriteString(AddressPrefix)
	buf.WriteString(encodedKey)
	buf.WriteString(encodedChecksum)
	return buf.String()
}

// Verify reports whether the given signature is valid for the given data.
func (a Address) Verify(data []byte, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(a), data, signature)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (a Address) MarshalText() (text []byte, err error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (a *Address) UnmarshalText(text []byte) error {
	addr, err := ParseAddress(string(text))
	if err != nil {
		return err
	}

	*a = addr
	return nil
}
