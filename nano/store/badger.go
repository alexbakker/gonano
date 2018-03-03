package store

import (
	"errors"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
	"github.com/dgraph-io/badger"
)

const (
	idPrefixBlock byte = iota
	idPrefixAddress
	idPrefixFrontier
)

// BadgerStore represents a Nano block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	txn *badger.Txn
}

// NewBadgerStore initializes/opens a badger database at the given directory.
func NewBadgerStore(dir string) (*BadgerStore, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStore{db: db}, nil
}

// Close closes the database
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Purge purges any old/deleted keys from the database.
func (s *BadgerStore) Purge() error {
	if err := s.db.PurgeOlderVersions(); err != nil {
		return err
	}

	return s.db.RunValueLogGC(0.5)
}

func (s *BadgerStore) View(fn func(txn StoreTxn) error) error {
	return s.db.View(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn})
	})
}

func (s *BadgerStore) Update(fn func(txn StoreTxn) error) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn})
	})
}

// SetGenesis initializes the database with the given genesis block, if needed.
func (t *BadgerStoreTxn) SetGenesis(genesis *block.OpenBlock) error {
	empty, err := t.Empty()
	if err != nil {
		return err
	}

	if !empty {
		if _, err := t.GetBlock(genesis.Hash()); err != nil {
			return ErrBadGenesis
		}
	} else {
		return t.AddBlock(genesis)
	}

	return nil
}

// Empty reports whether the database is empty or not.
func (t *BadgerStoreTxn) Empty() (bool, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte{idPrefixBlock}
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		return false, nil
	}

	return true, nil
}

// AddBlock adds the given block to the database.
func (t *BadgerStoreTxn) AddBlock(blk block.Block) error {
	hash := blk.Hash()
	blockBytes, err := blk.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + block.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])

	// never overwrite blocks implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrBlockExists
	}

	return t.txn.SetWithMeta(key[:], blockBytes, blk.ID())
}

// GetBlock retrieves the block with the given hash from the database.
func (t *BadgerStoreTxn) GetBlock(hash block.Hash) (block.Block, error) {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	blockType := item.UserMeta()
	blockBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	blk, err := block.New(blockType)
	if err != nil {
		return nil, err
	}

	if err := blk.UnmarshalBinary(blockBytes); err != nil {
		return nil, err
	}

	return blk, nil
}

// HasBlock reports whether the database contains a block with the given hash.
func (t *BadgerStoreTxn) HasBlock(hash block.Hash) (bool, error) {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])

	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CountBlocks returns the total amount of blocks in the database.
func (t *BadgerStoreTxn) CountBlocks() (uint64, error) {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte{idPrefixBlock}
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		count++
	}

	return count, nil
}

func (t *BadgerStoreTxn) AddAddress(address wallet.Address, blk block.OpenBlock) error {
	blockBytes, err := blk.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + wallet.AddressSize]byte
	key[0] = idPrefixAddress
	copy(key[1:], address)

	// never overwrite addresses implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return errors.New("address already exists")
	}

	return t.txn.SetWithMeta(key[:], blockBytes, blk.ID())
}

func (t *BadgerStoreTxn) AddFrontier(frontier *block.Frontier) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], frontier.Hash[:])

	// allow overwriting frontiers implicitly
	return t.txn.Set(key[:], frontier.Address)
}

func (t *BadgerStoreTxn) GetFrontier(hash block.Hash) (*block.Frontier, error) {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	address, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	return &block.Frontier{Address: address, Hash: hash}, nil
}

func (t *BadgerStoreTxn) GetFrontiers() ([]*block.Frontier, error) {
	var frontiers []*block.Frontier
	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte{idPrefixFrontier}
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		address, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		var frontier block.Frontier
		frontier.Address = address
		copy(frontier.Hash[:], item.Key())

		frontiers = append(frontiers, &frontier)
	}

	return frontiers, nil
}

func (t *BadgerStoreTxn) DeleteFrontier(hash block.Hash) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return t.txn.Delete(key[:])
}

func (t *BadgerStoreTxn) CountFrontiers() (uint64, error) {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte{idPrefixFrontier}
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		count++
	}

	return count, nil
}
