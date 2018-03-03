package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/alexbakker/gonano/nano/wallet"
)

var (
	cores = runtime.NumCPU()
)

type result struct {
	Seed       *wallet.Seed
	Address    wallet.Address
	Iterations uint64
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

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("error: no prefix specified\n")
		return
	}

	prefix := os.Args[1]
	for _, c := range prefix {
		if !strings.ContainsRune(wallet.AddressEncodingAlphabet, c) {
			fmt.Printf("error: char '%c' is not in nano's encoding alphabet\n", c)
			return
		}
	}

	c := make(chan *result)
	for n := 0; n < cores; n++ {
		go findAddress(prefix, c)
	}

	fmt.Printf("searching for address with prefix '%s' on %d cores\n", prefix, cores)

	if res := <-c; res != nil {
		fmt.Printf("found a match! (after %d iterations)\n", res.Iterations)
		fmt.Printf("seed: %s\n", res.Seed)
		fmt.Printf("address: %s\n", res.Address)
	} else {
		fmt.Printf("no match found!\n")
	}
}
