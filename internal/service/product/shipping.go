package product

// Full list of supported countries **and** the broader regions they belong to.
// This slice can be used for validation, autocomplete, or display purposes.
var supportedLocations = []string{
	// ── Regions ───────────────────────────────────────────────────────
	"European Union", "Europe", "North America", "Central America",
	"South America", "Caribbean", "Oceania", "Asia", "Africa",

	// ── European Union (27 members) ─────────────────────────────────
	"Austria", "Belgium", "Bulgaria", "Croatia", "Cyprus", "Czech Republic",
	"Denmark", "Estonia", "Finland", "France", "Germany", "Greece", "Hungary",
	"Ireland", "Italy", "Latvia", "Lithuania", "Luxembourg", "Malta",
	"Netherlands", "Poland", "Portugal", "Romania", "Slovakia", "Slovenia",
	"Spain", "Sweden",

	// ── Europe (non‑EU) ───────────────────────────────────────────────
	"Albania", "Andorra", "Bosnia and Herzegovina", "Georgia", "Iceland",
	"Liechtenstein", "Moldova", "Monaco", "Montenegro", "North Macedonia",
	"Norway", "Russia", "San Marino", "Serbia", "Switzerland", "Turkey",
	"Ukraine", "United Kingdom", "Vatican City",

	// ── North America ─────────────────────────────────────────────────
	"United States", "Canada", "Mexico",

	// ── Central America ───────────────────────────────────────────────
	"Belize", "Costa Rica", "El Salvador", "Guatemala", "Honduras",
	"Nicaragua", "Panama",

	// ── South America ─────────────────────────────────────────────────
	"Argentina", "Bolivia", "Brazil", "Chile", "Colombia", "Ecuador",
	"Paraguay", "Peru", "Uruguay", "Venezuela",

	// ── Caribbean ─────────────────────────────────────────────────────
	"Bahamas", "Barbados", "Cuba", "Dominican Republic", "Jamaica",
	"Puerto Rico", "Trinidad and Tobago",

	// ── Oceania ───────────────────────────────────────────────────────
	"Australia", "New Zealand", "Fiji", "Papua New Guinea", "Samoa",

	// ── Asia (selected) ───────────────────────────────────────────────
	"Japan", "South Korea", "Thailand", "Malaysia", "Israel",

	// ── Africa (selected) ─────────────────────────────────────────────
	"South Africa", "Nigeria", "Kenya", "Egypt", "Morocco", "Ghana",
	"Tunisia", "Algeria",
}

var slMap = map[string]bool{}

func init() {
	for _, location := range supportedLocations {
		slMap[location] = true
	}
}

func SupportedLocations() []string {
	return supportedLocations
}

func IsLocationSupported(location string) bool {
	_, ok := slMap[location]
	return ok
}
