package wallet

import (
	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/crypto/ed25519"
)

// Account represents an account in a Nano wallet.
type Account struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
}

// NewAccount creates a new account with the given private key.
func NewAccount(key ed25519.PrivateKey) *Account {
	return &Account{
		pubKey:  key.Public().(ed25519.PublicKey),
		privKey: key,
	}
}

// Address returns the public key of this account as an Address type.
func (a *Account) Address() nano.Address {
	var address nano.Address
	copy(address[:], a.pubKey)
	return address
}

func (a *Account) Sign(hash block.Hash) block.Signature {
	var sig block.Signature
	copy(sig[:], ed25519.Sign(a.privKey, hash[:]))
	return sig
}
