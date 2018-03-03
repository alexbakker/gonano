package main

import (
	"net"
	"os"
	"os/user"
	"path"

	"github.com/alexbakker/gonano/nano/node"
	"github.com/alexbakker/gonano/nano/store"
	"github.com/alexbakker/gonano/nano/store/genesis"
)

func resolve(address string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}
	return addr
}

func prepareDir() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	dir := path.Join(user.HomeDir, ".config/gonano/db")
	if err = os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}

	return dir
}

func main() {
	ledgerOpts := store.LedgerOptions{
		GenesisBlock:   genesis.LiveBlock,
		GenesisBalance: genesis.LiveBalance,
	}

	nodeOpts := node.DefaultOptions
	nodeOpts.Peers = []*net.UDPAddr{
		resolve("rai.raiblocks.net:7075"),
		//resolve("192.168.1.60:7075"),
	}

	// create gonano config directory
	prepareDir()

	// open the database
	db, err := store.NewBadgerStore("/home/alex/.config/gonano/db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// initialize the ledger
	ledger, err := store.NewLedger(db, ledgerOpts)
	if err != nil {
		panic(err)
	}

	// start up the node
	nanode, err := node.New(ledger, nodeOpts)
	if err != nil {
		panic(err)
	}

	panic(nanode.Run())
}
