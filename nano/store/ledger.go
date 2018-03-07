package store

import (
	"errors"
	"fmt"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	ErrBadWork         = errors.New("bad work")
	ErrBadGenesis      = errors.New("genesis block in store doesn't match the given block")
	ErrMissingPrevious = errors.New("previous block does not exist")
	ErrMissingSource   = errors.New("source block does not exist")
)

type Ledger struct {
	opts LedgerOptions
	db   Store
}

type LedgerOptions struct {
	GenesisBlock   *block.OpenBlock
	GenesisBalance wallet.Balance
}

func NewLedger(store Store, opts LedgerOptions) (*Ledger, error) {
	ledger := Ledger{opts: opts, db: store}

	// initialize the store with the genesis block if needed
	if err := ledger.setGenesis(opts.GenesisBlock, opts.GenesisBalance); err != nil {
		return nil, err
	}

	return &ledger, nil
}

func (l *Ledger) setGenesis(blk *block.OpenBlock, balance wallet.Balance) error {
	hash := blk.Hash()

	// is the work valid?
	if !blk.Valid() {
		fmt.Printf("bad work for genesis block")
	}

	// is the signature valid?
	signature := blk.Signature()
	if !blk.Address.Verify(hash[:], signature[:]) {
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

	// is the signature of this block valid?
	signature := blk.Signature()
	if !blk.Address.Verify(hash[:], signature[:]) {
		return errors.New("bad block signature")
	}

	// does this account already exist?
	_, err := txn.GetAddress(blk.Address)
	if err == nil {
		return errors.New("account already exists")
	}

	// obtain the pending transaction info
	pending, err := txn.GetPending(blk.Address, blk.SourceHash)
	if err != nil {
		return err
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

	// todo: update representative voting weight

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

	// is the hash of the previous block a frontier?
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		// todo: this indicates a fork!
		return err
	}

	// is the signature of this block valid?
	signature := blk.Signature()
	if !frontier.Address.Verify(hash[:], signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if !info.HeadBlock.Equal(frontier.Hash) {
		return errors.New("unexpected head block for account")
	}

	// make sure this is not a negative or zero spend
	comp := blk.Balance.Compare(info.Balance)
	if comp == wallet.BalanceCompBigger || comp == wallet.BalanceCompEqual {
		return fmt.Errorf("negative spend: %s >= %s", blk.Balance, info.Balance)
	}

	// todo: update representative voting weight

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

	// update the frontier of this account
	if err := txn.DeleteFrontier(hash); err != nil {
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

	// is the hash of the previous block a frontier?
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		// todo: this indicates a fork!
		return err
	}

	// is the signature of this block valid?
	signature := blk.Signature()
	if !frontier.Address.Verify(hash[:], signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if !info.HeadBlock.Equal(frontier.Hash) {
		return errors.New("unexpected head block for account")
	}

	// obtain the pending transaction info
	pending, err := txn.GetPending(frontier.Address, blk.SourceHash)
	if err != nil {
		return err
	}

	// todo: update representative voting weight

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

	// update the frontier of this account
	if err := txn.DeleteFrontier(hash); err != nil {
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

	// is the hash of the previous block a frontier?
	frontier, err := txn.GetFrontier(blk.Root())
	if err != nil {
		// todo: this indicates a fork!
		return err
	}

	// is the signature of this block valid?
	signature := blk.Signature()
	if !frontier.Address.Verify(hash[:], signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information and do some sanity checks
	info, err := txn.GetAddress(frontier.Address)
	if err != nil {
		return err
	}
	if !info.HeadBlock.Equal(frontier.Hash) {
		return errors.New("unexpected head block for account")
	}

	// update the address info
	info.HeadBlock = hash
	info.RepBlock = hash
	if err := txn.UpdateAddress(frontier.Address, info); err != nil {
		return err
	}

	// todo: update representative voting weight

	// update the frontier of this account
	if err := txn.DeleteFrontier(hash); err != nil {
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

	// is the work valid?
	if !blk.Valid() {
		return ErrBadWork
	}

	// is this block hash unique?
	found, err := txn.HasBlock(hash)
	if err != nil {
		return err
	}
	if found {
		return ErrBlockExists
	}

	// does the root block hash exist?
	found, err = txn.HasBlock(blk.Root())
	if err != nil {
		return err
	}
	if !found {
		return ErrMissingPrevious
	}

	switch b := blk.(type) {
	case *block.OpenBlock:
		return l.addOpenBlock(txn, b)
	case *block.SendBlock:
		return l.addSendBlock(txn, b)
	case *block.ReceiveBlock:
		return l.addReceiveBlock(txn, b)
	case *block.ChangeBlock:
		return l.addChangeBlock(txn, b)
	default:
		panic("bad block type")
	}
}

func (l *Ledger) AddBlock(blk block.Block) error {
	return l.db.Update(func(txn StoreTxn) error {
		return l.addBlock(txn, blk)
	})
}

func (l *Ledger) AddBlocks(blocks []block.Block) error {
	return l.db.Update(func(txn StoreTxn) error {
		for _, blk := range blocks {
			if err := l.addBlock(txn, blk); err != nil {
				if err != ErrMissingPrevious && err != ErrMissingSource {
					// todo: add to unchecked list
					fmt.Printf("error adding block %s: %s\n", blk.Hash(), err)
				}
				continue
			} else {
				fmt.Printf("added block: %s\n", blk.Hash())
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
