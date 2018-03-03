package util

import (
	"bytes"
	"testing"
)

func TestReverseBytes(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !bytes.Equal(ReverseBytes(data), []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}) {
		t.Fatal("somehow someone managed to mess up the reverse function")
	}
}
