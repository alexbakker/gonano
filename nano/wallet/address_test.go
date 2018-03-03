package wallet

import "testing"

func TestWalletAddress(t *testing.T) {
	s := "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3"

	address, err := ParseAddress(s)
	if err != nil {
		t.Fatal(err)
	}

	if address.String() != s {
		t.Fatalf("addresses are not equal")
	}
}
