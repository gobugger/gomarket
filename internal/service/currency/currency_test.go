package currency

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

func TestConversions(t *testing.T) {
	ps = map[Currency]float64{
		"USD": 307.66,
		"EUR": 264.50,
	}

	require.Equal(t, int64(30766), cryptoPrice("USD"))
	require.Equal(t, int64(26450), cryptoPrice("EUR"))

	require.Equal(t, XMR(), Fiat2Crypto("USD", 30766))
	require.Equal(t, XMR(), Fiat2Crypto("EUR", 26450))

	require.Equal(t, int64(30766), Crypto2Fiat("USD", XMR()))

	require.Equal(t, int64(30766), Fiat2Fiat("EUR", "USD", 26450))

	require.Equal(t, XMR(), Fiat2Crypto("USD", Crypto2Fiat("USD", XMR())))
	require.Equal(t, XMR(), Fiat2Crypto("EUR", Crypto2Fiat("EUR", XMR())))

	eur := int64(2000000)
	usd := Fiat2Fiat("EUR", "USD", eur)
	require.Equal(t, eur, Fiat2Fiat("USD", "EUR", usd))
}

func TestFiat2Fiat(t *testing.T) {
	for range 100 {
		amount := rand.Int63n(1000000000)
		// From eur -> usd -> eur
		require.InDelta(t, amount, Fiat2Fiat("USD", "EUR", Fiat2Fiat("EUR", "USD", amount)), 1)
		// usd -> eur -> usd
		require.InDelta(t, amount, Fiat2Fiat("EUR", "USD", Fiat2Fiat("USD", "EUR", amount)), 1)
	}
}

func TestDisplayFiat(t *testing.T) {
	tests := []struct {
		amount   int64
		expected string
	}{
		{amount: 100, expected: "1.00"},
		{amount: 500, expected: "5.00"},
		{amount: 1000, expected: "10.00"},
		{amount: 1, expected: "0.01"},
		{amount: 10, expected: "0.10"},
		{amount: 1513, expected: "15.13"},
	}

	for _, test := range tests {
		actual := DisplayFiat(test.amount)
		require.Equal(t, test.expected, actual)
	}
}

func TestRaw2Whole2Raw(t *testing.T) {
	require.Equal(t, big.NewFloat(1.0).SetPrec(prec), Raw2Whole(XMR()))
	require.Equal(t, XMR(), Whole2Raw(big.NewFloat(1)))
}

func TestRaw2Decimal(t *testing.T) {
	require.Equal(t, "1", Raw2Decimal(XMR()))
	require.Equal(t, "0.123", Raw2Decimal(big.NewInt(123000000000)))
	require.Equal(t, "0.123006", Raw2Decimal(big.NewInt(123006000000)))
}
