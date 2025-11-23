package currency

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestDisplay(t *testing.T) {
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

func TestConversions(t *testing.T) {
	ps = map[Currency]float64{
		"USD": 307.66,
		"EUR": 264.50,
	}

	require.Equal(t, int64(30766), xmrPrice("USD"))
	require.Equal(t, int64(26450), xmrPrice("EUR"))
	require.Equal(t, XMR, Fiat2XMR("USD", 30766))
	require.Equal(t, XMR, Fiat2XMR("EUR", 26450))
	require.Equal(t, int64(30766), Fiat2Fiat("EUR", "USD", 26450))
	require.Equal(t, XMR, Fiat2XMR("USD", XMR2Fiat("USD", XMR)))
	require.Equal(t, XMR, Fiat2XMR("EUR", XMR2Fiat("EUR", XMR)))

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
