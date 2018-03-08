package store

import (
	"errors"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	ErrBlockExists = errors.New("block already exists")
	ErrStoreEmpty  = errors.New("the store is empty")
)

// Store is an interface that all Nano block lattice stores need to implement.
type Store interface {
	Close() error
	Purge() error
	View(fn func(txn StoreTxn) error) error
	Update(fn func(txn StoreTxn) error) error
}

type StoreTxn interface {
	Empty() (bool, error)
	AddBlock(blk block.Block) error
	GetBlock(hash block.Hash) (block.Block, error)
	DeleteBlock(hash block.Hash) error
	HasBlock(hash block.Hash) (bool, error)
	CountBlocks() (uint64, error)
	AddAddress(address wallet.Address, info *AddressInfo) error
	GetAddress(address wallet.Address) (*AddressInfo, error)
	UpdateAddress(address wallet.Address, info *AddressInfo) error
	DeleteAddress(address wallet.Address) error
	AddFrontier(frontier *block.Frontier) error
	GetFrontier(hash block.Hash) (*block.Frontier, error)
	GetFrontiers() ([]*block.Frontier, error)
	DeleteFrontier(hash block.Hash) error
	CountFrontiers() (uint64, error)
	AddPending(destination wallet.Address, hash block.Hash, pending *Pending) error
	GetPending(destination wallet.Address, hash block.Hash) (*Pending, error)
	DeletePending(destination wallet.Address, hash block.Hash) error
	AddRepresentation(address wallet.Address, amount wallet.Balance) error
	SubRepresentation(address wallet.Address, amount wallet.Balance) error
	GetRepresentation(address wallet.Address) (wallet.Balance, error)
}
