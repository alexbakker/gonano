# Nano Protocol

This document describes the network protocol used by the Nano cryptocurrency. It
is licensed under CC BY-SA.

At the time of writing, there is no official protocol documentation. As such,
the source code of [the C++
implementation](https://github.com/nanocurrency/raiblocks), which is also
essentially undocumented, was studied to obtain information about the protocol.
Naming has been kept similar to the C++ implementation to avoid confusion.

Be sure to read and understand the whitepaper first:
https://nano.org/en/whitepaper. The basic concepts of Nano will not be covered
here.

## Networking

Nano nodes use two internet protocols for communication, UDP and TCP.

UDP is used for ping packets, broadcasting new blocks and voting. Because ping
packets are sent regularly to other peers, hole punching will occur. Forwarding
a port for this is usually not necessary for clients.

TCP is used when large transfers need to occur. Examples of large transfers are
frontier and block transfers. Clients will usually not serve on this port,
unless they have their port forwarded.

Listeners for both of these protocols are assumed to be bound to the same port.
The default port is 7075.

## Cryptographic primitives

Nano uses a number of cryptographic primitives to facilitate hashing and
signing. 

For hashing, BLAKE2b is used. 

For signing, a modified version of Ed25519 is used. It uses BLAKE2b as the
hashing algorithm instead of SHA-512.

## Packets

The C++ implementation doesn't appear to take endianness into account. This
means that the encoding of integers is platform-dependent. The platforms that
Nano currently runs on are all little endian, so we'll assume that all integers
are encoded in little endian for now. The exception to this rule are balances,
those are encoded in big endian.

Every packet starts with a header.

| Length | Contents                    |
| :----- | :-------------------------- |
| `1`    | 'R' encoded in ASCII        |
| `1`    | Network ID encoded in ASCII |
| `1`    | `uint8_t` VersionMax        |
| `1`    | `uint8_t` VersionMax        |
| `1`    | `uint8_t` VersionMin        |
| `1`    | `uint8_t` MessageType       |
| `2`    | `uint16_t` Extensions       |

Network identifiers:

| Value | Name |
| :---- | :--- |
| `A`   | Test |
| `B`   | Beta |
| `C`   | Live |

After the header, the message content follows. 

### Messages

There are a number of message types.

| Value  | Name             | UDP  | TCP  |
| :----- | :--------------- | :--- | :--- |
| `0x00` | Invalid          | ✓    | ✓    |
| `0x01` | Not a type       | ✓    | ✓    |
| `0x02` | Keep alive       | ✓    | ✗    |
| `0x03` | Publish          | ✓    | ✗    |
| `0x04` | Confirm Req      | ✓    | ✗    |
| `0x05` | Confirm ACK      | ✓    | ✗    |
| `0x06` | Bulk Pull        | ✗    | ✓    |
| `0x07` | Bulk Push        | ✗    | ✓    |
| `0x08` | Frontier Req     | ✗    | ✓    |
| `0x09` | Bulk Pull Blocks | ✗    | ✓    |

#### Keep alive

Nodes send keep alive packets to eachother on a regular basis. This packet is
used to share a random selection of their peer list. It is also used as a ping
mechanism to make sure a node is still up and running.

| Length | Contents                |
| :----- | :---------------------- |
| `?`    | List of peers (up to 8) |

A peer represented as an IPv6 address and a port.

| Length | Contents        |
| :----- | :-------------- |
| `16`   | IPv6 address    |
| `2`    | `uint16_t` Port |

__NOTE:__ The C++ implementation currently has a bug where it will do an out of
bounds read if less than 8 peers are in the packet. See also:
https://github.com/nanocurrency/raiblocks/issues/673. So for now, it's best to
always send 8 peers. If less than 8 peers are available, the rest of the list is
filled with unspecified ip's and zeroed ports.

In the C++ implementation the default interval is 60 seconds. After 5 minutes a
node is considered not alive anymore.

#### Publish

Publish packets are used to broadcast new blocks to nodes.

| Length | Contents |
| :----- | :------- |
| `?`    | Block    |

This packet contains a block. To learn more about what blocks look like, read
the Blocks chapter.

#### Confirm Req

When a node encounters a fork, it requests a vote using this packet type. 

| Length | Contents |
| :----- | :------- |
| `?`    | Block    |

This packet contains a block. To learn more about what blocks look like, read
the Blocks chapter.

#### Confirm ACK

This packet represents a vote.

| Length | Contents                  |
| :----- | :------------------------ |
| `32`   | Representative public key |
| `64`   | Signature                 |
| `8`    | `uint64_t` Sequence       |
| `?`    | Block                     |

#### Bulk Pull

| Length | Contents                  |
| :----- | :------------------------ |
| `32`   | Public key of the account |
| `32`   | End block hash            |

When this packet is sent to a node, it will respond with a stream of blocks.

To indicate the end of a transmission, a block with type: "Not a type" is sent.

#### Bulk Push

?

#### Frontier Req

| Length | Contents         |
| :----- | :--------------- |
| `32`   | Start public key |
| `4`    | `uint32_t` Age   |
| `4`    | `uint32_t` Count |

When this packet is sent to a node, it will respond with a stream of frontiers.

So, to retrieve all frontiers a node has, set the start public key to zero and
age and count to UINT32_MAX.

A frontier consists of the public key of an account and the newest block hash on
its chain. 

| Length | Contents           |
| :----- | :----------------- |
| `32`   | Account public key |
| `32`   | Newest block hash  |

To indicate the end of a transmission, a zeroed public key and block hash is
sent.

#### Bulk Pull Blocks

| Length | Contents           |
| :----- | :----------------- |
| `32`   | Minimum block hash |
| `32`   | Maximum block hash |
| `1`    | `uint8_t` Mode     |
| `4`    | `uint32_t` Count   |

There are two different modes. 

| Value  | Name     |
| :----- | :------- |
| `0x00` | List     |
| `0x01` | Checksum |

### Blocks

| Value  | Name       |
| :----- | :--------- |
| `0x00` | Invalid    |
| `0x01` | Not a type |
| `0x02` | Send       |
| `0x03` | Receive    |
| `0x04` | Open       |
| `0x05` | Change     |
| `0x06` | State      |

Every block contains a proof of work value. Read the Work chapter to learn more
about it.

Blocks also have a 256-bit BLAKE2b hash associated with them. The way those are
calculated differs per block type.

Block signatures are calculated by signing the block hash with the account
public key.

#### Send

A send block is created to transfer funds to another account.

| Length | Contents               |
| :----- | :--------------------- |
| `32`   | Previous block hash    |
| `32`   | Destination public key |
| `16`   | `uint128_t` Balance    |
| `64`   | Signature              |
| `8`    | `uint64_t` Work        |

The hash of this block is calculated by concatenating [Previous block hash,
Destination public key, Balance] and hashing the result with BLAKE2b.

#### Receive

A receive block is created to accept the transfer of funds that occurred in a
send block.

| Length | Contents            |
| :----- | :------------------ |
| `32`   | Previous block hash |
| `32`   | Source block hash   |
| `64`   | Signature           |
| `8`    | `uint64_t` Work     |

The hash of this block is calculated by concatenating [Previous block hash,
Source block hash] and hashing the result with BLAKE2b.

#### Open

An open block is created to open an account.

| Length | Contents                  |
| :----- | :------------------------ |
| `32`   | Source block hash         |
| `32`   | Representative public key |
| `32`   | Account public key        |
| `64`   | Signature                 |
| `8`    | `uint64_t` Work           |

The hash of this block is calculated by concatenating [Source block hash,
Representative public key, Account public key] and hashing the result with
BLAKE2b.

#### Change

A change block is created when an account wants to change its representative.

| Length | Contents                  |
| :----- | :------------------------ |
| `32`   | Previous block hash       |
| `32`   | Representative public key |
| `64`   | Signature                 |
| `8`    | `uint64_t` Work           |

The hash of this block is calculated by concatenating [Previous block hash,
Representative public key] and hashing the result with BLAKE2b.

#### State

State blocks (previously known as universal blocks) are meant to replace all
other block types.

| Length | Contents                  |
| :----- | :------------------------ |
| `32`   | Account public key        |
| `32`   | Previous block hash       |
| `32`   | Representative public key |
| `16`   | `uint128_t` Balance       |
| `32`   | Link                      |
| `64`   | Signature                 |
| `8`    | `uint64_t` Work           |

__NOTE:__ For this block type, the work value is encoded in big endian.

The hash of this block is calculated by concatenating [32 byte preamble (value:
0x06), Account public key, Previous block hash, Representative public key,
Balance, Link] and hashing the result with BLAKE2b.

### Work

To protect the network from spam, nodes need to perform some proof of work
before they can create a valid block. 

A 64-bit counter and the previous block hash are concatenated and hashed using
BLAKE2b to create a 8-byte hash. If the resulting hash is below the threshold of
0xffffffc000000000, the counter is incremented and the process is repeated until
the threshold is exceeded.

### Balance

Balances are represented as a 128-bit unsigned integer. They are encoded in big
endian.
