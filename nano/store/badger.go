package store

import (
	"errors"
	"os"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
)

const (
	idPrefixBlock byte = iota
	idPrefixUncheckedBlockPrevious
	idPrefixUncheckedBlockSource
	idPrefixAddress
	idPrefixFrontier
	idPrefixPending
	idPrefixRepresentation
)

const (
	badgerMaxOps = 10000
)

// BadgerStore represents a Nano block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	txn *badger.Txn
	db  *badger.DB
	ops uint64
}

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore(dir string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(dir)
	opts.ValueLogLoadingMode = badgerOpts.FileIO

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.Mkdir(dir, 0700); err != nil {
			return nil, err
		}
	}

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

func (s *BadgerStore) View(fn func(txn StoreTxn) error) error {
	return s.db.View(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn: txn, db: s.db})
	})
}

func (s *BadgerStore) Update(fn func(txn StoreTxn) error) error {
	t := &BadgerStoreTxn{txn: s.db.NewTransaction(true), db: s.db}
	defer t.txn.Discard()

	if err := fn(t); err != nil {
		return err
	}

	return t.txn.Commit()
}

func (t *BadgerStoreTxn) set(key []byte, val []byte) error {
	if err := t.txn.Set(key, val); err != nil {
		return err
	}

	t.ops++
	return nil
}

func (t *BadgerStoreTxn) setWithMeta(key []byte, val []byte, meta byte) error {
	if err := t.txn.SetEntry(badger.NewEntry(key, val).WithMeta(meta)); err != nil {
		return err
	}

	t.ops++
	return nil
}

func (t *BadgerStoreTxn) delete(key []byte) error {
	if err := t.txn.Delete(key); err != nil {
		return err
	}

	t.ops++
	return nil
}

// Empty reports whether the database is empty or not.
func (t *BadgerStoreTxn) Empty() (bool, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		return false, nil
	}

	return true, nil
}

func (t *BadgerStoreTxn) Flush() error {
	if t.ops >= badgerMaxOps {
		if err := t.txn.Commit(); err != nil {
			return err
		}

		t.ops = 0
		t.txn = t.db.NewTransaction(true)
	}

	return nil
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

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrBlockExists
	}

	return t.txn.SetEntry(badger.NewEntry(key[:], blockBytes).WithMeta(blk.ID()))
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

func (t *BadgerStoreTxn) DeleteBlock(hash block.Hash) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return t.delete(key[:])
}

// AddUncheckedBlock adds the given block to the database.
func (t *BadgerStoreTxn) AddUncheckedBlock(parentHash block.Hash, blk block.Block, kind UncheckedKind) error {
	blockBytes, err := blk.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + block.HashSize]byte
	key[0] = uncheckedKindToPrefix(kind)
	copy(key[1:], parentHash[:])

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrBlockExists
	}

	return t.txn.SetEntry(badger.NewEntry(key[:], blockBytes).WithMeta(blk.ID()))
}

// GetUncheckedBlock retrieves the block with the given hash from the database.
func (t *BadgerStoreTxn) GetUncheckedBlock(parentHash block.Hash, kind UncheckedKind) (block.Block, error) {
	var key [1 + block.HashSize]byte
	key[0] = uncheckedKindToPrefix(kind)
	copy(key[1:], parentHash[:])

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

func (t *BadgerStoreTxn) DeleteUncheckedBlock(parentHash block.Hash, kind UncheckedKind) error {
	var key [1 + block.HashSize]byte
	key[0] = uncheckedKindToPrefix(kind)
	copy(key[1:], parentHash[:])
	return t.delete(key[:])
}

// HasUncheckedBlock reports whether the database contains a block with the given hash.
func (t *BadgerStoreTxn) HasUncheckedBlock(hash block.Hash, kind UncheckedKind) (bool, error) {
	var key [1 + block.HashSize]byte
	key[0] = uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])

	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (t *BadgerStoreTxn) walkUncheckedBlocks(kind UncheckedKind, visit UncheckedBlockWalkFunc) error {
	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := [...]byte{uncheckedKindToPrefix(kind)}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		item := it.Item()

		blockType := item.UserMeta()
		blockBytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		blk, err := block.New(blockType)
		if err != nil {
			return err
		}

		if err := blk.UnmarshalBinary(blockBytes); err != nil {
			return err
		}

		if err := visit(blk, kind); err != nil {
			return err
		}
	}

	return nil
}

func (t *BadgerStoreTxn) WalkUncheckedBlocks(visit UncheckedBlockWalkFunc) error {
	var err error
	if err = t.walkUncheckedBlocks(UncheckedKindPrevious, visit); err != nil {
		return err
	}

	return t.walkUncheckedBlocks(UncheckedKindSource, visit)
}

func (t *BadgerStoreTxn) countUncheckedBlocks(kind UncheckedKind) uint64 {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{uncheckedKindToPrefix(kind)}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}

	return count
}

func (t *BadgerStoreTxn) CountUncheckedBlocks() (uint64, error) {
	return t.countUncheckedBlocks(UncheckedKindPrevious) +
		t.countUncheckedBlocks(UncheckedKindSource), nil
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

	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}

	return count, nil
}

func (t *BadgerStoreTxn) AddAddress(address nano.Address, info *AddressInfo) error {
	infoBytes, err := info.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + nano.AddressSize]byte
	key[0] = idPrefixAddress
	copy(key[1:], address[:])

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return errors.New("address already exists")
	}

	return t.set(key[:], infoBytes)
}

func (t *BadgerStoreTxn) GetAddress(address nano.Address) (*AddressInfo, error) {
	var key [1 + nano.AddressSize]byte
	key[0] = idPrefixAddress
	copy(key[1:], address[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	infoBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	var info AddressInfo
	if err := info.UnmarshalBinary(infoBytes); err != nil {
		return nil, err
	}

	return &info, nil
}

func (t *BadgerStoreTxn) UpdateAddress(address nano.Address, info *AddressInfo) error {
	infoBytes, err := info.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + nano.AddressSize]byte
	key[0] = idPrefixAddress
	copy(key[1:], address[:])

	return t.set(key[:], infoBytes)
}

func (t *BadgerStoreTxn) DeleteAddress(address nano.Address) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixAddress
	copy(key[1:], address[:])
	return t.delete(key[:])
}

func (t *BadgerStoreTxn) HasAddress(address nano.Address) (bool, error) {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], address[:])

	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (t *BadgerStoreTxn) AddFrontier(frontier *block.Frontier) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], frontier.Hash[:])

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return errors.New("frontier already exists")
	}

	return t.set(key[:], frontier.Address[:])
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

	frontier := block.Frontier{Hash: hash}
	copy(frontier.Address[:], address)
	return &frontier, nil
}

func (t *BadgerStoreTxn) GetFrontiers() ([]*block.Frontier, error) {
	var frontiers []*block.Frontier
	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := [...]byte{idPrefixFrontier}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		item := it.Item()
		address, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		var frontier block.Frontier
		copy(frontier.Address[:], address)
		copy(frontier.Hash[:], item.Key())

		frontiers = append(frontiers, &frontier)
	}

	return frontiers, nil
}

func (t *BadgerStoreTxn) DeleteFrontier(hash block.Hash) error {
	var key [1 + block.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return t.delete(key[:])
}

func (t *BadgerStoreTxn) CountFrontiers() (uint64, error) {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{idPrefixFrontier}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}

	return count, nil
}

func (t *BadgerStoreTxn) AddPending(destination nano.Address, hash block.Hash, pending *Pending) error {
	pendingBytes, err := pending.MarshalBinary()
	if err != nil {
		return err
	}

	var key [1 + PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+nano.AddressSize:], hash[:])

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return errors.New("pending transaction already exists")
	}

	return t.set(key[:], pendingBytes)
}

func (t *BadgerStoreTxn) GetPending(destination nano.Address, hash block.Hash) (*Pending, error) {
	var key [1 + PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+nano.AddressSize:], hash[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	pendingBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	var pending Pending
	if err := pending.UnmarshalBinary(pendingBytes); err != nil {
		return nil, err
	}

	return &pending, nil
}

func (t *BadgerStoreTxn) DeletePending(destination nano.Address, hash block.Hash) error {
	var key [1 + PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+nano.AddressSize:], hash[:])
	return t.delete(key[:])
}

func (t *BadgerStoreTxn) setRepresentation(address nano.Address, amount nano.Balance) error {
	var key [1 + nano.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])

	amountBytes, err := amount.MarshalBinary()
	if err != nil {
		return err
	}

	return t.set(key[:], amountBytes)
}

func (t *BadgerStoreTxn) AddRepresentation(address nano.Address, amount nano.Balance) error {
	oldAmount, err := t.GetRepresentation(address)
	if err != nil {
		return err
	}

	return t.setRepresentation(address, oldAmount.Add(amount))
}

func (t *BadgerStoreTxn) SubRepresentation(address nano.Address, amount nano.Balance) error {
	oldAmount, err := t.GetRepresentation(address)
	if err != nil {
		return err
	}

	return t.setRepresentation(address, oldAmount.Sub(amount))
}

func (t *BadgerStoreTxn) GetRepresentation(address nano.Address) (nano.Balance, error) {
	var key [1 + nano.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nano.ZeroBalance, nil
		}
		return nano.ZeroBalance, err
	}

	amountBytes, err := item.ValueCopy(nil)
	if err != nil {
		return nano.ZeroBalance, err
	}

	var amount nano.Balance
	if err := amount.UnmarshalBinary(amountBytes); err != nil {
		return nano.ZeroBalance, err
	}

	return amount, nil
}
