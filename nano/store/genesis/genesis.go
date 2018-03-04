package genesis

import (
	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	LiveBlock = &block.OpenBlock{
		SourceHash:     util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Representative: util.MustDecodeHex("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Address:        util.MustDecodeHex("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
		Common: block.CommonBlock{
			Work:      0x62f05417dd3fb691,
			Signature: util.MustDecodeHex64("9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"),
		},
	}
	LiveBalance = wallet.ParseBalanceInts(0xffffffffffffffff, 0xffffffffffffffff)
)
