package block

import (
	"encoding/hex"
	"testing"
)

func TestBlockWork(t *testing.T) {
	work := Work(0xc2c306caf73b836f)
	hash := mustDecodeHash(t, "6529C605D4016F486B60861C49DDAD128D77642E748B3FE13BE411F00BA0918B")
	if !work.Valid(hash) {
		t.Errorf("work not valid")
	}

	worker := NewWorker(work, hash)
	if worker.Generate() != work {
		t.Fatal("work not equal")
	}

	worker.Reset()
	worker.Generate()
}

func mustDecodeHash(t *testing.T, s string) Hash {
	var hash Hash
	bytes, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	copy(hash[:], bytes)
	return hash
}
