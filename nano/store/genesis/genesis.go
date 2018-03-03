package genesis

import (
	"encoding/hex"

	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	LiveBlock = &block.OpenBlock{
		SourceHash:     mustDecodeHash("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Representative: mustDecodeAddress("xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3"),
		Address:        mustDecodeAddress("xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3"),
		Common: block.CommonBlock{
			Work:      0x62f05417dd3fb691,
			Signature: mustDecodeSignature("9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"),
		},
	}
	LiveBalance = wallet.ParseBalanceInts(0xffffffffffffffff, 0xffffffffffffffff)
)

func mustDecodeHex(s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bytes
}

func mustDecodeHash(s string) block.Hash {
	var hash block.Hash
	bytes := mustDecodeHex(s)
	copy(hash[:], bytes)
	return hash
}

func mustDecodeSignature(s string) block.Signature {
	var signature block.Signature
	bytes := mustDecodeHex(s)
	copy(signature[:], bytes)
	return signature
}

func mustDecodeAddress(s string) wallet.Address {
	addr, err := wallet.ParseAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}
