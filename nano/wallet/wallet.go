package wallet

type Wallet struct {
	seed     *Seed
	accounts []*Account
	index    uint32
}

func New(seed *Seed, index uint32) (*Wallet, error) {
	accounts := []*Account{}
	for i := uint32(0); i < index+1; i++ {
		key, err := seed.Key(i)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, NewAccount(key))
	}

	return &Wallet{
		seed:     seed,
		accounts: accounts,
		index:    index,
	}, nil
}

func Generate() (*Wallet, error) {
	// generate a new seed for the wallet
	seed, err := GenerateSeed()
	if err != nil {
		return nil, err
	}

	return New(seed, 0)
}

func (w *Wallet) Accounts() []*Account {
	accounts := make([]*Account, len(w.accounts))
	copy(accounts, w.accounts)
	return accounts
}
