export GO15VENDOREXPERIMENT=1

all: nano-node nano-vanity nano-wallet

nano-node: prep
	go build -o build/bin/nano-node github.com/alexbakker/gonano/cmd/nano-node

nano-vanity: prep
	go build -o build/bin/nano-vanity github.com/alexbakker/gonano/cmd/nano-vanity

nano-wallet: prep
	go build -o build/bin/nano-wallet github.com/alexbakker/gonano/cmd/nano-wallet

test:
	GOCACHE=off go test -v $(shell go list ./... | grep -v vendor)

prep:
	mkdir -p build/bin

clean:
	rm -rf build

loc:
	find . -name "*.go" -not -path "./vendor/*" -not -path "./nano/internal/uint128/*" -not -path "./nano/crypto/ed25519/*" | xargs wc -l
