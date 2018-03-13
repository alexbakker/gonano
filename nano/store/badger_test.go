package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/crypto/random"
	"github.com/alexbakker/gonano/nano/wallet"
)

func generateBlock(t testing.TB) block.Block {
	blk := &block.OpenBlock{
		Representative: make(wallet.Address, wallet.AddressSize),
		Address:        make(wallet.Address, wallet.AddressSize),
	}
	random.Bytes(blk.Representative)
	random.Bytes(blk.Address)
	random.Bytes(blk.SourceHash[:])
	random.Bytes(blk.Common.Signature[:])
	return blk
}

func TestBadgerWrite(t *testing.T) {
	const n = 100000
	ledger := initTestLedger(t)
	store := ledger.store
	//defer ledger.Close(t)

	start := time.Now()
	err := store.Update(func(txn StoreTxn) error {
		for i := 0; i < n; i++ {
			blk := generateBlock(t)
			if err := txn.AddBlock(blk); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	end := time.Now()

	fmt.Printf("write benchmark: %d ns/op\n", end.Sub(start).Nanoseconds()/n)
}
