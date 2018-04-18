package nano

import (
	"testing"
)

func TestNanoBalance(t *testing.T) {
	b1 := ParseBalanceInts(0xffffffffffffffff, 0xffffffffffffffff)
	b2 := ParseBalanceInts(0, 1)
	b1Units := map[string]string{
		"raw":  "340282366920938463463374607431768211455",
		"uxrb": "340282366920938463463.374607431768211455",
		"mxrb": "340282366920938463.463374607431768211455",
		"xrb":  "340282366920938.463463374607431768211455",
		"kxrb": "340282366920.938463463374607431768211455",
		"Mxrb": "340282366.920938463463374607431768211455",
		"Gxrb": "340282.366920938463463374607431768211455",
	}
	b2Units := map[string]string{
		"raw":  "1",
		"uxrb": "0.000000000000000001",
		"mxrb": "0.000000000000000000001",
		"xrb":  "0.000000000000000000000001",
		"kxrb": "0.000000000000000000000000001",
		"Mxrb": "0.000000000000000000000000000001",
		"Gxrb": "0.000000000000000000000000000000001",
	}
	b1TruncatedUnits := map[string]string{
		"raw":  "340282366920938463463374607431768211455",
		"uxrb": "340282366920938463463.374607",
		"mxrb": "340282366920938463.463374",
		"xrb":  "340282366920938.463463",
		"kxrb": "340282366920.938463",
		"Mxrb": "340282366.920938",
		"Gxrb": "340282.36692",
	}

	compare := func(b Balance, m map[string]string, p int32) {
		for unit, s := range m {
			res := b.UnitString(unit, p)
			if res != s {
				t.Errorf("(%s) expected: %s, got: %s\n", unit, s, res)
			}
		}
	}

	compare(b1, b1Units, BalanceMaxPrecision)
	compare(b1, b1TruncatedUnits, 6)
	compare(b2, b2Units, BalanceMaxPrecision)

	if b1.String() != b1Units["Mxrb"] {
		t.Errorf("unexpected fmt.Stringer result")
	}

	for unit, s := range b1Units {
		b, err := ParseBalance(s, unit)
		if err != nil {
			t.Error(err)
			continue
		}

		if !b.Equal(b1) {
			t.Errorf("(%s) expected: %s, got: %s\n", unit, b1.UnitString(unit, BalanceMaxPrecision), b.UnitString(unit, BalanceMaxPrecision))
		}
	}

	for unit, s := range b1TruncatedUnits {
		b, err := ParseBalance(s, unit)
		if err != nil {
			t.Error(err)
			continue
		}

		res := b.UnitString(unit, 6)
		if res != s {
			t.Errorf("(%s) expected: %s, got: %s\n", unit, s, res)
		}
	}
}
