package block

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"

	"golang.org/x/crypto/blake2b"
)

const (
	workSize      = 8
	WorkThreshold = 0xffffffc000000000
)

type Work uint64

type Worker struct {
	root *Hash
	work Work
	hash hash.Hash
}

func (w Work) Valid(root Hash) bool {
	return NewWorker(w, root).Valid()
}

// MarshalText implements the encoding.TextMarshaler interface.
func (w Work) MarshalText() (text []byte, err error) {
	return []byte(w.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (w *Work) UnmarshalText(text []byte) error {
	size := hex.DecodedLen(len(text))
	if size != workSize {
		return fmt.Errorf("bad work size: %d", size)
	}

	var work [workSize]byte
	if _, err := hex.Decode(work[:], text); err != nil {
		return err
	}

	*w = Work(binary.BigEndian.Uint64(work[:]))
	return nil
}

// String implements the fmt.Stringer interface.
func (w Work) String() string {
	var bytes [workSize]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(w))
	return hex.EncodeToString(bytes[:])
}

func NewWorker(work Work, root Hash) *Worker {
	hash, err := blake2b.New(workSize, nil)
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
	var workBytes [workSize]byte
	binary.LittleEndian.PutUint64(workBytes[:], uint64(w.work))

	w.hash.Reset()
	w.hash.Write(workBytes[:])
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
