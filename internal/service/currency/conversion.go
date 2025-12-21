package currency

import (
	"fmt"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/shopspring/decimal"
	"log/slog"
	"math"
)

func whole() decimal.Decimal {
	if config.Cryptocurrency == "NANO" {
		return nano
	} else {
		return xmr
	}
}

func Fiat2Crypto(c Currency, amount int64) decimal.Decimal {
	y := whole().Mul(decimal.NewFromInt(amount))
	return y.Div(decimal.NewFromInt(cryptoPrice(c))).Truncate(0)
}

func Crypto2Fiat(c Currency, amount decimal.Decimal) int64 {
	p := decimal.NewFromInt(cryptoPrice(c))
	result := p.Mul(Raw2Whole(amount))
	return result.BigInt().Int64()
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
func Whole2Raw(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(whole())
}

// Convert from piconeros to decimal XMR
func Raw2Whole(raw decimal.Decimal) decimal.Decimal {
	return raw.Div(whole())
}

func Raw2Decimal(raw decimal.Decimal) string {
	return Raw2Whole(raw).String()
}
