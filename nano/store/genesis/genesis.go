package genesis

import (
	"fmt"

	"github.com/alexbakker/gonano/nano"
	"github.com/alexbakker/gonano/nano/block"
	"github.com/alexbakker/gonano/nano/internal/util"
	"github.com/alexbakker/gonano/nano/node/proto"
)

type Genesis struct {
	Block         block.OpenBlock
	Balance       nano.Balance
	WorkThreshold uint64
}

var (
	Live = Genesis{
		Block: block.OpenBlock{
			SourceHash:     util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
			Representative: util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
			Address:        util.MustDecodeHex32("e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba"),
			Work:           0x62f05417dd3fb691,
			Signature:      util.MustDecodeHex64("9f0c933c8ade004d808ea1985fa746a7e95ba2a38f867640f53ec8f180bdfe9e2c1268dead7c2664f356e37aba362bc58e46dba03e523a7b5a19e4b6eb12bb02"),
		},
		Balance:       nano.ParseBalanceInts(0xffffffffffffffff, 0xffffffffffffffff),
		WorkThreshold: uint64(0xffffffc000000000),
	}

	Beta = Genesis{
		Block: block.OpenBlock{
			SourceHash:     util.MustDecodeHex32("a59a47cc4f593e75ae9ad653fda9358e2f7898d9acc8c60e80d0495ce20fba9f"),
			Representative: util.MustDecodeHex32("a59a47cc4f593e75ae9ad653fda9358e2f7898d9acc8c60e80d0495ce20fba9f"),
			Address:        util.MustDecodeHex32("a59a47cc4f593e75ae9ad653fda9358e2f7898d9acc8c60e80d0495ce20fba9f"),
			Work:           0x000000000f0aaeeb,
			Signature:      util.MustDecodeHex64("a726490e3325e4fa59c1c900d5b6eebb15fe13d99f49d475b93f0aacc5635929a0614cf3892764a04d1c6732a0d716ffeb254d4154c6f544d11e6630f201450b"),
		},
		Balance:       nano.ParseBalanceInts(0xffffffffffffffff, 0xffffffffffffffff),
		WorkThreshold: uint64(0xffffffc000000000),
	}
)

func Get(network proto.Network) (Genesis, error) {
	switch network {
	case proto.NetworkLive:
		return Live, nil
	case proto.NetworkBeta:
		return Beta, nil
	}

	return Genesis{}, fmt.Errorf("unsupported network: %s", network)
}
