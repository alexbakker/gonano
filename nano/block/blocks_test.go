package block

import (
	"testing"

	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	openBlock = &OpenBlock{
		SourceHash:     util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Representative: util.MustDecodeHex("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Address:        util.MustDecodeHex("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Common: CommonBlock{
			Work:      0x62f05417dd3fb691,
			Signature: util.MustDecodeHex64("9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"),
		},
	}
	sendBlock = &SendBlock{
		PreviousHash: util.MustDecodeHex32("4270F4FB3A820FE81827065F967A9589DF5CA860443F812D21ECE964AC359E05"),
		Destination:  util.MustDecodeHex("0000000000000000000000000000000000000000000000000000000000000000"),
		Balance:      wallet.ParseBalanceInts(0, 0),
		Common: CommonBlock{
			Work:      0x7202df8a7c380578,
			Signature: util.MustDecodeHex64("047115CB577AC78F5C66AD79BBF47540DE97A441456004190F22025FE4255285F57010D962601AE64C266C98FA22973DD95AC62309634940B727AC69F0C86D03"),
		},
	}
	receiveBlock = &ReceiveBlock{
		PreviousHash: util.MustDecodeHex32("80D6C93EE26F64353B5A69B3CE641F8DE244FDCC79798502EB43D8FDEFC93976"),
		SourceHash:   util.MustDecodeHex32("1036E1153620383ADCAF49AB7C5F27C07233D98DF573A4F19CE75DBCC7F5AD9B"),
		Common: CommonBlock{
			Work:      0xf0d8f54f62e0e757,
			Signature: util.MustDecodeHex64("BBE8BA0AB966E209EA1F661023B76F3DF453F0714A96BE2DA5E717A78DCDF7C59D83BCE96FE3285BBE2EF151C6CCCBFEE0C758DAD25956A64F73E2D6C8366F0B"),
		},
	}
	changeBlock = &ChangeBlock{
		PreviousHash:   util.MustDecodeHex32("3abd41f575184a02b28f98f1c71684c8f6adc0f7334eb32e8232cdad65609e23"),
		Representative: util.MustDecodeHex("7d5c67cf17432c5c88fa739a3cc88f894da3eec7b977a804780977cae35fa5a8"),
		Common: CommonBlock{
			Work:      0x8f48c0b00946163c,
			Signature: util.MustDecodeHex64("1009a0c2fbc189dc41d13daa9d7a6e1ab2d6c4e06200aeca1b7ae0c27bf454b2c60f446df0d69a877d3a98e3128ab3442058172fbd965024519630ace93b670e"),
		},
	}
)

func TestBlockOpenMarshal(t *testing.T) {
	bytes, err := openBlock.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	var blk OpenBlock
	if err = blk.UnmarshalBinary(bytes); err != nil {
		t.Fatal(err)
	}

	if blk.Hash() != openBlock.Hash() || blk.Common != openBlock.Common {
		t.Fatalf("blocks not equal")
	}
}

func TestBlockSendMarshal(t *testing.T) {
	bytes, err := sendBlock.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	var blk SendBlock
	if err = blk.UnmarshalBinary(bytes); err != nil {
		t.Fatal(err)
	}

	if blk.Hash() != sendBlock.Hash() || blk.Common != sendBlock.Common {
		t.Fatalf("blocks not equal")
	}
}

func TestBlockReceiveMarshal(t *testing.T) {
	bytes, err := receiveBlock.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	var blk ReceiveBlock
	if err = blk.UnmarshalBinary(bytes); err != nil {
		t.Fatal(err)
	}

	if blk.Hash() != receiveBlock.Hash() || blk.Common != receiveBlock.Common {
		t.Fatalf("blocks not equal")
	}
}

func TestBlockChangeMarshal(t *testing.T) {
	bytes, err := changeBlock.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	var blk ChangeBlock
	if err = blk.UnmarshalBinary(bytes); err != nil {
		t.Fatal(err)
	}

	if blk.Hash() != changeBlock.Hash() || blk.Common != changeBlock.Common {
		t.Fatalf("blocks not equal")
	}
}
