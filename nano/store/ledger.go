package store

import (
	"errors"
	"fmt"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
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
	// initialize the store with the genesis block if needed
	err := store.Update(func(txn StoreTxn) error {
		return txn.SetGenesis(opts.GenesisBlock)
	})

	if err != nil {
		return nil, err
	}

	return &Ledger{opts: opts, db: store}, nil
}

/*func (l *Ledger) GetBlock(hash block.Hash) (block.Block, error) {
	return l.db.GetBlock(hash)
}

func (l *Ledger) HasBlock(hash block.Hash) (bool, error) {
	return l.db.HasBlock(hash)
}*/

func (l *Ledger) addOpenBlock(txn StoreTxn, blk *block.OpenBlock) error {
	hash := blk.Hash()

	// is the signature of this block valid?
	signature := blk.Signature()
	if !blk.Address.Verify(hash[:], signature[:]) {
		return errors.New("bad block signature")
	}

	// does this account already exist?

	return nil
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

	return nil
}

func (l *Ledger) addReceiveBlock(txn StoreTxn, blk *block.ReceiveBlock) error {
	//hash := blk.Hash()

	return nil
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

	return nil
}

func (l *Ledger) addBlock(txn StoreTxn, blk block.Block) error {
	hash := blk.Hash()

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
		// todo: add to unchecked list
		return errors.New("previous block does not exist")
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
				fmt.Printf("error adding block: %s\n", err)
				continue
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
