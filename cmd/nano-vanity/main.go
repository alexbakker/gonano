package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/alexbakker/gonano/nano/wallet"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "nano-vanity",
		Short: "A terribly inefficient vanity address generator for Nano",
		Run:   startVanity,
	}

	prefix  string
	threads int
)

type result struct {
	Seed       *wallet.Seed
	Address    wallet.Address
	Iterations uint64
}

func init() {
	rootCmd.Flags().StringVar(&prefix, "prefix", "", "prefix to search for")
	rootCmd.Flags().IntVar(&threads, "threads", runtime.NumCPU(), "number of threads to use")
	rootCmd.MarkFlagRequired("prefix")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}

func startVanity(cmd *cobra.Command, args []string) {
	for _, c := range prefix {
		if !strings.ContainsRune(wallet.AddressEncodingAlphabet, c) {
			fmt.Printf("error: char '%c' is not in nano's encoding alphabet\n", c)
			return
		}
	}

	c := make(chan *result)
	for n := 0; n < threads; n++ {
		go findAddress(prefix, c)
	}

	fmt.Printf("searching for address with prefix '%s' on %d threads\n", prefix, threads)

	if res := <-c; res != nil {
		fmt.Printf("found a match! (after %d iterations)\n", res.Iterations)
		fmt.Printf("seed: %s\n", res.Seed)
		fmt.Printf("address: %s\n", res.Address)
	} else {
		fmt.Printf("no match found!\n")
	}
}

func findAddress(prefix string, c chan *result) {
	for i := uint64(0); ; i++ {
		seed, err := wallet.GenerateSeed()
		if err != nil {
			panic(err)
		}

		key, err := seed.Key(0)
		if err != nil {
			panic(err)
		}

		addr := wallet.NewAccount(key).Address()
		if strings.HasPrefix(addr.String()[5:], prefix) {
			c <- &result{seed, addr, i}
		}
	}
}
