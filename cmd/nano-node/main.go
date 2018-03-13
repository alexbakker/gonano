package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"

	"github.com/alexbakker/gonano/nano/node"
	"github.com/alexbakker/gonano/nano/store"
	"github.com/alexbakker/gonano/nano/store/genesis"
)

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
	dir := prepareDir()

	// open the database
	db, err := store.NewBadgerStore(dir)
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
	go func() {
		if err := nanode.Run(); err != nil {
			panic(err)
		}
	}()

	// handle any signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)
	switch <-sigc {
	case syscall.SIGTERM, syscall.SIGINT:
		fmt.Println()
		fmt.Println("exit signal caught")

		fmt.Println("stopping the node")
		if err := nanode.Stop(); err != nil {
			fmt.Printf("error stopping node: %s\n", err)
		}

		fmt.Println("closing badger database")
		if err := db.Close(); err != nil {
			fmt.Printf("error closing db: %s\n", err)
		}

		os.Exit(0)

	case syscall.SIGHUP:
		// reload config
	}
}

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
