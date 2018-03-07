package store

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/store/genesis"
)

type testLedger struct {
	*Ledger
	store Store
	dir   string
}

func initTestLedger(t *testing.T) *testLedger {
	dir, err := ioutil.TempDir("", "gonano_test_")
	if err != nil {
		t.Fatal(err)
	}

	store, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	ledger, err := NewLedger(store, LedgerOptions{
		GenesisBlock:   genesis.LiveBlock,
		GenesisBalance: genesis.LiveBalance,
	})
	if err != nil {
		t.Fatal(err)
	}

	return &testLedger{
		Ledger: ledger,
		store:  store,
		dir:    dir,
	}
}

func (l *testLedger) Close(t *testing.T) {
	if err := l.store.Close(); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(l.dir); err != nil {
		t.Fatal(err)
	}
}

func parseBlocks(t *testing.T, filename string) (blocks []block.Block) {
	type fileStruct struct {
		Blocks []json.RawMessage `json:"blocks"`
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var file fileStruct
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}

	for _, data := range file.Blocks {
		var values map[string]interface{}
		if err = json.Unmarshal(data, &values); err != nil {
			t.Fatal(err)
		}

		id, ok := values["type"]
		if !ok {
			t.Fatalf("no 'type' key found in block")
		}

		var blk block.Block
		switch id {
		case "send":
			blk = new(block.SendBlock)
		case "receive":
			blk = new(block.ReceiveBlock)
		case "open":
			blk = new(block.OpenBlock)
		case "change":
			blk = new(block.ChangeBlock)
		default:
			t.Fatalf("unsupported block type: %s", id)
		}

		if err = json.Unmarshal(data, blk); err != nil {
			t.Fatal(err)
		}

		blocks = append(blocks, blk)
	}

	return
}

func TestLedgerBlocks(t *testing.T) {
	ledger := initTestLedger(t)
	defer ledger.Close(t)

	blocks := parseBlocks(t, "./testdata/blocks.json")
	for _, blk := range blocks {
		if err := ledger.AddBlock(blk); err != nil {
			t.Fatal(err)
		}
	}
}
