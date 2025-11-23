package currency

var currencySymbols = map[Currency]string{
	// --- European currencies ---
	"EUR": "€",  // Euro
	"GBP": "£",  // British Pound
	"CHF": "Fr", // Swiss Franc
	"SEK": "kr", // Swedish Krona
	"NOK": "kr", // Norwegian Krone
	"DKK": "kr", // Danish Krone
	"PLN": "zł", // Polish Zloty
	"CZK": "Kč", // Czech Koruna
	"HUF": "Ft", // Hungarian Forint

	// --- Other major currencies ---
	"USD": "$",   // US Dollar
	"CAD": "C$",  // Canadian Dollar
	"AUD": "A$",  // Australian Dollar
	"NZD": "NZ$", // New Zealand Dollar
	"JPY": "¥",   // Japanese Yen
	"CNY": "¥",   // Chinese Yuan
	"HKD": "HK$", // Hong Kong Dollar
	"SGD": "S$",  // Singapore Dollar
	"KRW": "₩",   // South Korean Won
	"INR": "₹",   // Indian Rupee
	"BRL": "R$",  // Brazilian Real
	"MXN": "$",   // Mexican Peso
	"ZAR": "R",   // South African Rand
}
