package store

import (
	"errors"
	"fmt"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/store/genesis"
)

var (
	ErrBadWork         = errors.New("bad work")
	ErrBadGenesis      = errors.New("genesis block in store doesn't match the given block")
	ErrMissingPrevious = errors.New("previous block does not exist")
	ErrMissingSource   = errors.New("source block does not exist")
	ErrUnchecked       = errors.New("block was added to the unchecked list")
	ErrFork            = errors.New("a fork was detected")
	ErrNotFound        = errors.New("item not found in the store")
)

type Ledger struct {
	opts LedgerOptions
	db   Store
}

type LedgerOptions struct {
	Genesis genesis.Genesis
}

func NewLedger(store Store, opts LedgerOptions) (*Ledger, error) {
	ledger := Ledger{opts: opts, db: store}

	// initialize the store with the genesis block if needed
	if err := ledger.setGenesis(&opts.Genesis.Block, opts.Genesis.Balance); err != nil {
		return nil, err
	}

	return &ledger, nil
}

func (l *Ledger) setGenesis(blk *block.OpenBlock, balance nano.Balance) error {
	hash := blk.Hash()

	// make sure the work value is valid
	if !blk.Valid(l.opts.Genesis.WorkThreshold) {
		return errors.New("bad work for genesis block")
	}

	// make sure the signature of this block is valid
	if !blk.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad signature for genesis block")
	}

	return l.db.Update(func(txn StoreTxn) error {
		empty, err := txn.Empty()
		if err != nil {
			return err
		}

		if !empty {
			// if the database is not empty, check if it has the same genesis
			// block as the one in the given options
			found, err := txn.HasBlock(hash)
			if err != nil {
				return err
			}
			if !found {
				return ErrBadGenesis
			}
		} else {
			if err := txn.AddBlock(blk); err != nil {
				return err
			}

			info := AddressInfo{
				HeadBlock: hash,
				RepBlock:  hash,
				OpenBlock: hash,
				Balance:   balance,
			}
			if err := txn.AddAddress(blk.Address, &info); err != nil {
				return err
			}

			return txn.AddFrontier(&block.Frontier{
				Address: blk.Address,
				Hash:    hash,
			})
		}

		return nil
	})
}

func (l *Ledger) addOpenBlock(txn StoreTxn, blk *block.OpenBlock) error {
	hash := blk.Hash()

	// make sure the signature of this block is valid
	if !blk.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad block signature")
	}

	// make sure this address doesn't already exist
	_, err := txn.GetAddress(blk.Address)
	if err == nil {
		return ErrFork
	}

	// obtain the pending transaction info
	pending, err := txn.GetPending(blk.Address, blk.SourceHash)
	if err != nil {
		return ErrMissingSource
	}

	// add address info
	info := AddressInfo{
		HeadBlock: hash,
		RepBlock:  hash,
		OpenBlock: hash,
		Balance:   pending.Amount,
	}
	if err := txn.AddAddress(blk.Address, &info); err != nil {
		return err
	}

	// delete the pending transaction
	if err := txn.DeletePending(blk.Address, blk.SourceHash); err != nil {
		return err
	}

	// update representative voting weight
	if err := txn.AddRepresentation(blk.Representative, pending.Amount); err != nil {
		return err
	}

	// add a frontier for this address
	frontier := block.Frontier{
		Address: blk.Address,
		Hash:    hash,
	}
	if err := txn.AddFrontier(&frontier); err != nil {
		return err
	}

	// finally, add the block
	return txn.AddBlock(blk)
}

func (l *Ledger) addSendBlock(txn StoreTxn, blk *block.SendBlock) error {
	hash := blk.Hash()

	// make sure the hash of the previous block is a frontier
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		return ErrFork
	}

	// make sure the signature of this block is valid
	if !frontier.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if info.HeadBlock != frontier.Hash {
		return errors.New("unexpected head block for account")
	}

	// make sure this is not a negative spend
	// (apparently zero spends are allowed?)
	if blk.Balance.Compare(info.Balance) == nano.BalanceCompBigger {
		return fmt.Errorf("negative spend: %s > %s", blk.Balance, info.Balance)
	}

	// add this to the pending transaction list
	pending := Pending{
		Address: frontier.Address,
		Amount:  info.Balance.Sub(blk.Balance),
	}
	if err := txn.AddPending(blk.Destination, hash, &pending); err != nil {
		return err
	}

	// update the address info
	info.HeadBlock = hash
	info.Balance = blk.Balance
	if err := txn.UpdateAddress(frontier.Address, info); err != nil {
		return err
	}

	// update representative voting weight
	rep, err := l.getRepresentative(txn, frontier.Address)
	if err != nil {
		return err
	}
	if err := txn.SubRepresentation(rep, blk.Balance); err != nil {
		return err
	}

	// update the frontier of this account
	if err := txn.DeleteFrontier(frontier.Hash); err != nil {
		return err
	}
	frontier = &block.Frontier{
		Address: frontier.Address,
		Hash:    hash,
	}
	if err := txn.AddFrontier(frontier); err != nil {
		return err
	}

	// finally, add the block to the store
	return txn.AddBlock(blk)
}

func (l *Ledger) addReceiveBlock(txn StoreTxn, blk *block.ReceiveBlock) error {
	hash := blk.Hash()

	// make sure the hash of the previous block is a frontier
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		return ErrFork
	}

	// make sure the signature of this block is valid
	if !frontier.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if info.HeadBlock != frontier.Hash {
		return errors.New("unexpected head block for account")
	}

	// obtain the pending transaction info
	pending, err := txn.GetPending(frontier.Address, blk.SourceHash)
	if err != nil {
		return ErrMissingSource
	}

	// update the address info
	info.HeadBlock = hash
	info.Balance = info.Balance.Add(pending.Amount)
	if err := txn.UpdateAddress(frontier.Address, info); err != nil {
		return err
	}

	// delete the pending transaction
	if err := txn.DeletePending(frontier.Address, blk.SourceHash); err != nil {
		return err
	}

	// update representative voting weight
	rep, err := l.getRepresentative(txn, frontier.Address)
	if err != nil {
		return err
	}
	if err := txn.AddRepresentation(rep, pending.Amount); err != nil {
		return err
	}

	// update the frontier of this account
	if err := txn.DeleteFrontier(frontier.Hash); err != nil {
		return err
	}
	frontier = &block.Frontier{
		Address: frontier.Address,
		Hash:    hash,
	}
	if err := txn.AddFrontier(frontier); err != nil {
		return err
	}

	// finally, add the block to the store
	return txn.AddBlock(blk)
}

func (l *Ledger) addChangeBlock(txn StoreTxn, blk *block.ChangeBlock) error {
	hash := blk.Hash()

	// make sure the hash of the previous block is a frontier
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		return ErrFork
	}

	// make sure the signature of this block is valid
	if !frontier.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if info.HeadBlock != frontier.Hash {
		return errors.New("unexpected head block for account")
	}

	// obtain the old representative
	oldRep, err := l.getRepresentative(txn, frontier.Address)
	if err != nil {
		return err
	}

	// update the address info
	info.HeadBlock = hash
	info.RepBlock = hash
	if err := txn.UpdateAddress(frontier.Address, info); err != nil {
		return err
	}

	// update representative voting weight
	if err := txn.SubRepresentation(oldRep, info.Balance); err != nil {
		return err
	}
	if err := txn.AddRepresentation(blk.Representative, info.Balance); err != nil {
		return err
	}

	// update the frontier of this account
	if err := txn.DeleteFrontier(frontier.Hash); err != nil {
		return err
	}
	frontier = &block.Frontier{
		Address: frontier.Address,
		Hash:    hash,
	}
	if err := txn.AddFrontier(frontier); err != nil {
		return err
	}

	// finally, add the block
	return txn.AddBlock(blk)
}

func (l *Ledger) addBlock(txn StoreTxn, blk block.Block) error {
	hash := blk.Hash()

	// make sure the work value is valid
	if !blk.Valid(l.opts.Genesis.WorkThreshold) {
		return ErrBadWork
	}

	// make sure the hash of this block doesn't exist yet
	found, err := txn.HasBlock(hash)
	if err != nil {
		return err
	}
	if found {
		return ErrBlockExists
	}

	// make sure the previous/source block exists
	found, err = txn.HasBlock(blk.Root())
	if err != nil {
		return err
	}
	if !found {
		switch b := blk.(type) {
		case *block.OpenBlock:
			return ErrMissingSource
		case *block.SendBlock:
			return ErrMissingPrevious
		case *block.ReceiveBlock:
			return ErrMissingPrevious
		case *block.ChangeBlock:
			return ErrMissingPrevious
		case *block.StateBlock:
			if b.IsOpen() {
				return ErrMissingSource
			} else {
				return ErrMissingPrevious
			}
		default:
			return block.ErrBadBlockType
		}
	}

	switch b := blk.(type) {
	case *block.OpenBlock:
		err = l.addOpenBlock(txn, b)
	case *block.SendBlock:
		err = l.addSendBlock(txn, b)
	case *block.ReceiveBlock:
		err = l.addReceiveBlock(txn, b)
	case *block.ChangeBlock:
		err = l.addChangeBlock(txn, b)
	default:
		return block.ErrBadBlockType
	}

	if err != nil {
		return err
	}

	// flush if needed
	return txn.Flush()
}

func (l *Ledger) addUncheckedBlock(txn StoreTxn, parentHash block.Hash, blk block.Block, kind UncheckedKind) error {
	found, err := txn.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	return txn.AddUncheckedBlock(parentHash, blk, kind)
}

func (l *Ledger) processUncheckedBlock(txn StoreTxn, blk block.Block, kind UncheckedKind) error {
	hash := blk.Hash()

	found, err := txn.HasUncheckedBlock(hash, kind)
	if err != nil {
		return err
	}

	if found {
		uncheckedBlk, err := txn.GetUncheckedBlock(hash, kind)
		if err != nil {
			return err
		}

		if err := l.processBlock(txn, uncheckedBlk); err == nil {
			// delete from the unchecked list if successful
			if err := txn.DeleteUncheckedBlock(hash, kind); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Ledger) processBlock(txn StoreTxn, blk block.Block) error {
	err := l.addBlock(txn, blk)

	switch err {
	case ErrMissingPrevious:
		// add to unchecked list
		if err := l.addUncheckedBlock(txn, blk.Root(), blk, UncheckedKindPrevious); err != nil {
			return err
		}
		return ErrUnchecked
	case ErrMissingSource:
		var source block.Hash
		switch b := blk.(type) {
		case *block.ReceiveBlock:
			source = b.SourceHash
		case *block.OpenBlock:
			source = b.SourceHash
		case *block.StateBlock:
			source = b.Link
		default:
			return errors.New("unexpected block type")
		}

		// add to unchecked list
		if err := l.addUncheckedBlock(txn, source, blk, UncheckedKindSource); err != nil {
			return err
		}

		return ErrUnchecked
	case nil:
		fmt.Printf("added block: %s\n", blk.Hash())

		// try to process any unchecked child blocks
		if err := l.processUncheckedBlock(txn, blk, UncheckedKindPrevious); err != nil {
			return err
		}

		if err := l.processUncheckedBlock(txn, blk, UncheckedKindSource); err != nil {
			return err
		}

		return nil
	case ErrBlockExists:
		// ignore
		return nil
	default:
		fmt.Printf("error adding block %s: %s\n", blk.Hash(), err)
		return err
	}
}

func (l *Ledger) AddBlock(blk block.Block) error {
	return l.db.Update(func(txn StoreTxn) error {
		err := l.processBlock(txn, blk)
		if err != nil && err != ErrUnchecked {
			fmt.Printf("try add err: %s\n", err)
		}

		return nil
	})
}

func (l *Ledger) AddBlocks(blocks []block.Block) error {
	return l.db.Update(func(txn StoreTxn) error {
		for _, blk := range blocks {
			err := l.processBlock(txn, blk)
			if err != nil && err != ErrUnchecked {
				fmt.Printf("try add err: %s\n", err)
			}
		}

		return nil
	})
}

func (l *Ledger) CountBlocks() (uint64, error) {
	var res uint64

	err := l.db.View(func(txn StoreTxn) error {
		count, err := txn.CountBlocks()
		if err != nil {
			return err
		}
		res = count
		return nil
	})

	return res, err
}

func (l *Ledger) CountUncheckedBlocks() (uint64, error) {
	var res uint64

	err := l.db.View(func(txn StoreTxn) error {
		count, err := txn.CountUncheckedBlocks()
		if err != nil {
			return err
		}
		res = count
		return nil
	})

	return res, err
}

func (l *Ledger) GetBalance(address nano.Address) (nano.Balance, error) {
	var balance nano.Balance

	err := l.db.View(func(txn StoreTxn) error {
		info, err := txn.GetAddress(address)
		if err != nil {
			return err
		}
		balance = info.Balance
		return nil
	})

	return balance, err
}

func (l *Ledger) GetFrontier(address nano.Address) (block.Hash, error) {
	var hash block.Hash

	err := l.db.View(func(txn StoreTxn) error {
		found, err := txn.HasAddress(address)
		if err != nil {
			return err
		}
		if !found {
			return ErrNotFound
		}

		info, err := txn.GetAddress(address)
		if err != nil {
			return nil
		}

		hash = info.HeadBlock
		return nil
	})

	return hash, err
}

func (l *Ledger) getRepresentative(txn StoreTxn, address nano.Address) (nano.Address, error) {
	info, err := txn.GetAddress(address)
	if err != nil {
		return nano.Address{}, err
	}

	blk, err := txn.GetBlock(info.RepBlock)
	if err != nil {
		return nano.Address{}, err
	}

	switch b := blk.(type) {
	case *block.OpenBlock:
		return b.Representative, nil
	case *block.ChangeBlock:
		return b.Representative, nil
	default:
		return nano.Address{}, errors.New("bad representative block type")
	}
}
