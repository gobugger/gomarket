package currency

import (
	"context"
	"encoding/json"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/shopspring/decimal"
	"log/slog"
	"sync"
	"time"
)

const (
	prec uint = 128
)

var (
	xmr  = decimal.NewFromInt(1e12)
	nano = decimal.NewFromInt(10).Pow(decimal.NewFromInt(30))
)

func XMR() decimal.Decimal {
	return xmr.Copy()
}

func NANO() decimal.Decimal {
	return nano.Copy()
}

func RAW() decimal.Decimal {
	return whole()
}

var ps = map[Currency]float64{}
var mtx sync.RWMutex

// Value of one Crypto in cents of currency c
func cryptoPrice(c Currency) int64 {
	if !c.IsSupported() {
		slog.Error("XMRPrice called with unsupported currency", "currency", c)
		c = DefaultCurrency
	}

	mtx.RLock()
	p, ok := ps[c]
	mtx.RUnlock()

	if !ok {
		slog.Error("Price not found for currency", "currency", c)
	}
	return int64(100 * p)
}

// Read prices from db on intervals
func Start(ctx context.Context, q *repo.Queries, interval time.Duration) error {
	if err := update(ctx, q); err != nil {
		return err
	}

	go func() {
		timer := time.NewTimer(0)
		quit := false
		for !quit {
			select {
			case <-ctx.Done():
				quit = true
			case <-timer.C:
				if err := update(ctx, q); err != nil {
					slog.Error("Failed to update prices", slog.Any("error", err))
				}
				timer.Reset(interval)
			}
		}
	}()

	return nil
}

// Set prices to db
func Set(ctx context.Context, q *repo.Queries, prices map[Currency]float64) error {
	data, err := json.Marshal(&prices)
	if err != nil {
		return err
	}

	return q.UpdateExchangeRates(ctx, data)
}

func update(ctx context.Context, q *repo.Queries) error {
	rates, err := q.GetExchangeRates(ctx)
	if err != nil {
		return err
	}

	prices := map[Currency]float64{}
	err = json.Unmarshal(rates.Data, &prices)

	mtx.Lock()
	ps = prices
	mtx.Unlock()
	return nil
}
