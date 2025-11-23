package currency

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
)

// Fiat's in cents (1/100) and XMR's in piconeros (1/1e12)

func Fiat2XMR(c Currency, amount int64) int64 {
	return int64(1e12 * amount / xmrPrice(c))
}

func XMR2Fiat(c Currency, amount int64) int64 {
	return int64(float64(xmrPrice(c)) * XMR2Float(amount))
}

func Fiat2Fiat(from Currency, to Currency, amount int64) int64 {
	if from == to {
		return amount
	}

	mtx.RLock()
	pf, ok := ps[from]
	pt, ok2 := ps[to]
	mtx.RUnlock()

	if !ok || !ok2 || pf == 0 {
		slog.Error("currency conversion failed", "from", from, "to", to)
		return 0.0
	}

	return int64(math.Round(float64(amount) * (pt / pf)))
}

func DisplayFiat(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount)/100)
}

// Convert form XMR to atomic unit AKA piconeros
func XMR2Int(xmr float64) int64 {
	return int64(xmr * 1e12)
}

// Convert from piconeros to decimal XMR
func XMR2Float(xmr int64) float64 {
	return float64(xmr) / 1e12
}

// Convert from piconeros to decimal string
// Copied from moneropay/walletrpc
func XMR2Decimal(xmr int64) string {
	if xmr == 0 {
		return "0"
	}
	str0 := fmt.Sprintf("%013d", xmr)
	l := len(str0)
	decimal := str0[:l-12]
	float := strings.TrimRight(str0[l-12:], "0")
	if len(float) == 0 {
		return decimal
	}
	return decimal + "." + float
}
