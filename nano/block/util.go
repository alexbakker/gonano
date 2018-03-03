package block

import "golang.org/x/crypto/blake2b"

func hashBytes(inputs ...[]byte) Hash {
	hash, err := blake2b.New(blake2b.Size256, nil)
	if err != nil {
		panic(err)
	}

	for _, data := range inputs {
		hash.Write(data)
	}

	var result Hash
	copy(result[:], hash.Sum(nil))
	return result
}
