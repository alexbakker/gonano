package block

import (
	"encoding/binary"
	"encoding/hex"
	"hash"

	"golang.org/x/crypto/blake2b"
)

const (
	WorkThreshold = 0xffffffc000000000
)

type Work uint64

type Worker struct {
	root *Hash
	work Work
	hash hash.Hash
}

// String implements the fmt.Stringer interface.
func (w Work) String() string {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(w))
	return hex.EncodeToString(bytes)
}

func (w Work) Valid(root Hash) bool {
	return NewWorker(w, root).Valid()
}

func NewWorker(work Work, root Hash) *Worker {
	hash, err := blake2b.New(8, nil)
	if err != nil {
		panic(err)
	}

	return &Worker{
		root: &root,
		work: work,
		hash: hash,
	}
}

func (w *Worker) Valid() bool {
	workBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(workBytes, uint64(w.work))

	w.hash.Reset()
	w.hash.Write(workBytes)
	w.hash.Write(w.root[:])

	sum := w.hash.Sum(nil)
	value := binary.LittleEndian.Uint64(sum)
	return value >= WorkThreshold
}

func (w *Worker) Generate() Work {
	for {
		if w.Valid() {
			return w.work
		}
		w.work++
	}
}

func (w *Worker) Reset() {
	w.work = 0
	w.hash.Reset()
}
