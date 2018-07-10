package block

import (
	"testing"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/internal/util"
)

var (
	openBlock = &OpenBlock{
		SourceHash:     util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Representative: util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Address:        util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Work:           0x62f05417dd3fb691,
		Signature:      util.MustDecodeHex64("9f0c933c8ade004d808ea1985fa746a7e95ba2a38f867640f53ec8f180bdfe9e2c1268dead7c2664f356e37aba362bc58e46dba03e523a7b5a19e4b6eb12bb02"),
	}
	sendBlock = &SendBlock{
		PreviousHash: util.MustDecodeHex32("4270f4fb3a820fe81827065f967a9589df5ca860443f812d21ece964ac359e05"),
		Destination:  util.MustDecodeHex32("0000000000000000000000000000000000000000000000000000000000000000"),
		Balance:      nano.ParseBalanceInts(0, 0),
		Work:         0x7202df8a7c380578,
		Signature:    util.MustDecodeHex64("047115cb577ac78f5c66ad79bbf47540de97a441456004190f22025fe4255285f57010d962601ae64c266c98fa22973dd95ac62309634940b727ac69f0c86d03"),
	}
	receiveBlock = &ReceiveBlock{
		PreviousHash: util.MustDecodeHex32("80d6c93ee26f64353b5a69b3ce641f8de244fdcc79798502eb43d8fdefc93976"),
		SourceHash:   util.MustDecodeHex32("1036e1153620383adcaf49ab7c5f27c07233d98df573a4f19ce75dbcc7f5ad9b"),
		Work:         0xf0d8f54f62e0e757,
		Signature:    util.MustDecodeHex64("bbe8ba0ab966e209ea1f661023b76f3df453f0714a96be2da5e717a78dcdf7c59d83bce96fe3285bbe2ef151c6cccbfee0c758dad25956a64f73e2d6c8366f0b"),
	}
	changeBlock = &ChangeBlock{
		PreviousHash:   util.MustDecodeHex32("3abd41f575184a02b28f98f1c71684c8f6adc0f7334eb32e8232cdad65609e23"),
		Representative: util.MustDecodeHex32("7d5c67cf17432c5c88fa739a3cc88f894da3eec7b977a804780977cae35fa5a8"),
		Work:           0x8f48c0b00946163c,
		Signature:      util.MustDecodeHex64("1009a0c2fbc189dc41d13daa9d7a6e1ab2d6c4e06200aeca1b7ae0c27bf454b2c60f446df0d69a877d3a98e3128ab3442058172fbd965024519630ace93b670e"),
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

	if blk.Hash() != openBlock.Hash() || blk.Signature != openBlock.Signature || blk.Work != openBlock.Work {
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

	if blk.Hash() != sendBlock.Hash() || blk.Signature != sendBlock.Signature || blk.Work != sendBlock.Work {
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

	if blk.Hash() != receiveBlock.Hash() || blk.Signature != receiveBlock.Signature || blk.Work != receiveBlock.Work {
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

	if blk.Hash() != changeBlock.Hash() || blk.Signature != changeBlock.Signature || blk.Work != changeBlock.Work {
		t.Fatalf("blocks not equal")
	}
}
