package store

import (
	"errors"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	ErrBlockExists = errors.New("block already exists")
	ErrStoreEmpty  = errors.New("the store is empty")
	ErrBadGenesis  = errors.New("genesis block in store doesn't match the given block")
)

// Store is an interface that all Nano block lattice stores need to implement.
type Store interface {
	Close() error
	Purge() error
	View(fn func(txn StoreTxn) error) error
	Update(fn func(txn StoreTxn) error) error
}

type StoreTxn interface {
	SetGenesis(genesis *block.OpenBlock) error
	AddBlock(blk block.Block) error
	GetBlock(hash block.Hash) (block.Block, error)
	HasBlock(hash block.Hash) (bool, error)
	CountBlocks() (uint64, error)
	AddAddress(address wallet.Address, blk block.OpenBlock) error
	AddFrontier(frontier *block.Frontier) error
	GetFrontier(hash block.Hash) (*block.Frontier, error)
	GetFrontiers() ([]*block.Frontier, error)
	DeleteFrontier(hash block.Hash) error
	CountFrontiers() (uint64, error)
}
