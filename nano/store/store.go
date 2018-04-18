package store

import (
	"errors"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
)

var (
	ErrBlockExists = errors.New("block already exists")
	ErrStoreEmpty  = errors.New("the store is empty")
)

type UncheckedKind byte

const (
	UncheckedKindPrevious UncheckedKind = iota
	UncheckedKindSource
)

// UncheckedBlockWalkFunc is the type of the function called for each unchecked
// block visited by WalkUncheckedBlocks.
type UncheckedBlockWalkFunc func(block block.Block, kind UncheckedKind) error

// Store is an interface that all Nano block lattice stores need to implement.
type Store interface {
	Close() error
	Purge() error
	View(fn func(txn StoreTxn) error) error
	Update(fn func(txn StoreTxn) error) error
}

type StoreTxn interface {
	Empty() (bool, error)
	Flush() error

	AddBlock(blk block.Block) error
	GetBlock(hash block.Hash) (block.Block, error)
	DeleteBlock(hash block.Hash) error
	HasBlock(hash block.Hash) (bool, error)
	CountBlocks() (uint64, error)

	AddUncheckedBlock(parentHash block.Hash, blk block.Block, kind UncheckedKind) error
	GetUncheckedBlock(parentHash block.Hash, kind UncheckedKind) (block.Block, error)
	DeleteUncheckedBlock(parentHash block.Hash, kind UncheckedKind) error
	HasUncheckedBlock(hash block.Hash, kind UncheckedKind) (bool, error)
	WalkUncheckedBlocks(visit UncheckedBlockWalkFunc) error
	CountUncheckedBlocks() (uint64, error)

	AddAddress(address nano.Address, info *AddressInfo) error
	GetAddress(address nano.Address) (*AddressInfo, error)
	UpdateAddress(address nano.Address, info *AddressInfo) error
	DeleteAddress(address nano.Address) error
	HasAddress(address nano.Address) (bool, error)

	AddFrontier(frontier *block.Frontier) error
	GetFrontier(hash block.Hash) (*block.Frontier, error)
	GetFrontiers() ([]*block.Frontier, error)
	DeleteFrontier(hash block.Hash) error
	CountFrontiers() (uint64, error)

	AddPending(destination nano.Address, hash block.Hash, pending *Pending) error
	GetPending(destination nano.Address, hash block.Hash) (*Pending, error)
	DeletePending(destination nano.Address, hash block.Hash) error

	AddRepresentation(address nano.Address, amount nano.Balance) error
	SubRepresentation(address nano.Address, amount nano.Balance) error
	GetRepresentation(address nano.Address) (nano.Balance, error)
}
