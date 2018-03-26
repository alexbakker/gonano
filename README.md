# gonano [![Build Status](https://travis-ci.org/alexbakker/gonano.svg?branch=master)](https://travis-ci.org/alexbakker/gonano)

__gonano__ is a __WIP__ implementation of the Nano cryptocurrency in Go.

This is a work in progress. Do not use this in production environments. All of
the exported API's are subject to change and should thus not be considered
stable. The same applies to the database format, configuration files and wallet
files.

Protocol documentation can be found at: [doc/protocol.md](doc/protocol.md).

The address of my motivational back account is:
xrb_1tt5p7agt63f3q37151o1yz3k1pfdan7wet9anejzrdqnuz5kgtcqiwwtfm6.

## Goals

The goals of this project are to:
- Provide an alternative to the existing C++ implementation
- Learn about the protocol
- Document the protocol
- Make Nano more accessible to developers

## Progress

- [ ] Network Protocol
  - [ ] Documentation (see also: [doc/protocol.md](doc/protocol.md))
  - [ ] Data structures
    - [x] Keep alive
    - [x] Publish
    - [x] Confirm Req
    - [x] Confirm ACK
    - [x] Bulk Pull
    - [ ] Bulk Push
    - [x] Frontier Req
    - [x] Bulk Pull Blocks
  - [ ] Synchronization
    - [x] Pull
    - [ ] Push
  - [x] Pinging
  - [ ] (Re)broadcasting blocks
  - [ ] Voting
- [x] Blocks
  - [x] Data structures
    - [x] Send
    - [x] Receive
    - [x] Open
    - [x] Change
    - [x] State
  - [x] Proof of Work
- [ ] Node
  - [ ] Persist state to a database
    - [x] Blocks
    - [x] Address balance
    - [x] Pending transactions
    - [x] Representatives (voting weight)
    - [ ] Votes
  - [x] Block verification
  - [ ] Fork resolution
    - [ ] Block rollback
  - [ ] RPC interface
- [ ] Wallet
  - [x] Data structures
    - [x] Balance
    - [x] Address
    - [x] Seed
  - [ ] Persist state to an encrypted file
  - [ ] RPC interface
- [ ] Tests

## Compiling

Go 1.8 or newer is required.

Run ``make all`` to build everything. Binaries can be found in the 'build'
folder.

Run ``make test`` to run the tests.

## Dependencies

This project directly depends on the following packages:
- [badger](https://github.com/dgraph-io/badger) - Fast key-value DB in Go
- [cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- [blake2b and ed25519](https://go.googlesource.com/crypto) - Go supplementary
  cryptography libraries
- [uint128](https://github.com/cockroachdb/cockroach/blob/master/pkg/util/uint128)
  128-bit unsigned integer package from CockroachDB
- [decimal](https://github.com/shopspring/decimal) - Arbitrary-precision
  fixed-point decimal numbers in go

The above packages are vendored and can be found in the vendor directory. The
ed25519 and uint128 packages are placed elsewhere as those had to be customized
for gonano.

## License

The source code of this project is licensed under the [MIT license](LICENSE).
The protocol documentation is licensed under [CC BY-SA](doc/LICENSE).
