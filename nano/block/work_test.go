package block

import (
	"encoding/hex"
	"testing"
)

func TestBlockWork(t *testing.T) {
	work := Work(0xc2c306caf73b836f)
	threshold := uint64(0xffffffc000000000)

	hash := mustDecodeHash(t, "6529c605d4016f486b60861c49ddad128d77642e748b3fe13be411f00ba0918b")
	if !work.Valid(hash, threshold) {
		t.Errorf("work not valid")
	}

	worker := NewWorker(work, hash, threshold)
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
