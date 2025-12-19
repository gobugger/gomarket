package currency

import (
	"fmt"
	"github.com/gobugger/gomarket/internal/config"
	"log/slog"
	"math"
	"math/big"
)

func whole() *big.Int {
	if config.Cryptocurrency == "NANO" {
		return nano
	} else {
		return xmr
	}
}

func Fiat2Crypto(c Currency, amount int64) *big.Int {
	y := new(big.Int).Mul(whole(), big.NewInt(amount))
	return y.Div(y, big.NewInt(cryptoPrice(c)))
}

func Crypto2Fiat(c Currency, amount *big.Int) int64 {
	p := new(big.Float).SetInt64(cryptoPrice(c))
	result, _ := p.Mul(p, Raw2Whole(amount)).Int(nil)
	return result.Int64()
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
func Whole2Raw(amount *big.Float) *big.Int {
	result, _ := new(big.Float).Mul(amount, new(big.Float).SetInt(whole())).Int(nil)
	return result
}

// Convert from piconeros to decimal XMR
func Raw2Whole(raw *big.Int) *big.Float {
	return new(big.Float).
		SetPrec(prec).
		SetRat(new(big.Rat).SetFrac(raw, whole()))
}

func Raw2Decimal(raw *big.Int) string {
	r := new(big.Rat).SetFrac(raw, whole())
	p, _ := r.FloatPrec()
	return r.FloatString(p)
}
