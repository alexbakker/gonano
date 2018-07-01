package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/alexbakker/gonano/cmd/nano-node/config"
	"github.com/alexbakker/gonano/nano/node"
	"github.com/alexbakker/gonano/nano/store"
	"github.com/alexbakker/gonano/nano/store/genesis"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "nano-node",
		Short: "Nano node",
		Run:   startNode,
	}

	man         *config.Manager
	cfg         config.Config
	cfgDefaults = config.Config{
		Addr: node.DefaultOptions.Address,
		Peers: []string{
			"rai.raiblocks.net:7075",
		},
	}

	logger = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
)

func main() {
	// set the umask of this process to 077
	// this ensures all written files are only readable/writable by the current user
	syscall.Umask(077)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error: %s", err.Error())
	}
}

func init() {
	rootCmd.Flags().StringVar(&cfg.Addr, "addr", cfgDefaults.Addr, "address to listen on for UDP and TCP")
	rootCmd.Flags().StringVar(&cfg.AddrRPC, "addr-rpc", cfgDefaults.AddrRPC, "address to listen on for RPC")
	rootCmd.Flags().StringVar(&cfg.AddrPprof, "addr-pprof", cfgDefaults.AddrPprof, "address to listen on for pprof")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	if man, err = config.NewManager(".config/gonano/node", "config.json"); err != nil {
		logger.Fatalf("config manager error: %s", err)
	}

	if err = man.Prepare(&cfgDefaults); err != nil {
		logger.Fatalf("config load error: %s", err)
	}

	if err = man.Load(&cfg); err != nil {
		logger.Fatalf("config load error: %s", err)
	}
}

func startNode(cmd *cobra.Command, args []string) {
	startPprof()

	ledgerOpts := store.LedgerOptions{
		GenesisBlock:   genesis.LiveBlock,
		GenesisBalance: genesis.LiveBalance,
	}
	nodeOpts := node.DefaultOptions
	nodeOpts.Peers = cfg.Peers
	nodeOpts.Address = cfg.Addr

	logger.Printf("opening badger database at %s", man.Dir())
	db, err := store.NewBadgerStore(path.Join(man.Dir(), "db"))
	if err != nil {
		logger.Fatalf("error opening database: %s", err)
	}
	defer db.Close()

	logger.Printf("initializing ledger")
	ledger, err := store.NewLedger(db, ledgerOpts)
	if err != nil {
		logger.Fatalf("error initializing ledger: %s", err)
	}

	logger.Printf("initializing node")
	nanode, err := node.New(ledger, nodeOpts)
	if err != nil {
		logger.Fatalf("error initializing node: %s", err)
	}
	go func() {
		logger.Printf("starting node")
		if err := nanode.Run(); err != nil {
			logger.Fatalf("error starting node: %s", err)
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		switch <-sigc {
		case syscall.SIGTERM, syscall.SIGINT:
			logger.Println("exit signal caught")

			logger.Println("stopping the node")
			if err := nanode.Stop(); err != nil {
				logger.Printf("error stopping node: %s", err)
			}

			logger.Println("closing badger database")
			if err := db.Close(); err != nil {
				logger.Printf("error closing db: %s", err)
			}

			os.Exit(0)
		case syscall.SIGHUP:
			logger.Println("error reloading config: not implemented")
		}
	}
}
