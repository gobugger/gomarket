package currency

import (
	"slices"
)

type Currency string

const (
	DefaultCurrency Currency = "USD"
)

var supportedCurrencies = []Currency{
	// --- European currencies ---
	"EUR", // Euro
	"GBP", // British Pound
	"CHF", // Swiss Franc
	"SEK", // Swedish Krona
	"NOK", // Norwegian Krone
	"DKK", // Danish Krone
	"PLN", // Polish Zloty
	"CZK", // Czech Koruna
	"HUF", // Hungarian Forint

	// --- Other major currencies ---
	"USD", // US Dollar
	"CAD", // Canadian Dollar
	"AUD", // Australian Dollar
	"NZD", // New Zealand Dollar
	"JPY", // Japanese Yen
	"CNY", // Chinese Yuan
	"HKD", // Hong Kong Dollar
	"SGD", // Singapore Dollar
	"KRW", // South Korean Won
	"INR", // Indian Rupee
	"BRL", // Brazilian Real
	"MXN", // Mexican Peso
	"ZAR", // South African Rand
}

func SupportedCurrencies() []Currency {
	out := make([]Currency, len(supportedCurrencies))
	copy(out, supportedCurrencies)
	return out
}

func (c Currency) IsSupported() bool {
	return slices.Contains(supportedCurrencies, c)
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) Symbol() string {
	if symbol, ok := currencySymbols[c]; ok {
		return symbol
	}
	return c.String()
}
