package currency

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"time"
)

func DebugStart(ctx context.Context, q *repo.Queries) {
	if err := Set(ctx, q, map[Currency]float64{"EUR": 280, "USD": 320}); err != nil {
		panic(err)
	}
	if err := Start(ctx, q, 15*time.Second); err != nil {
		panic(err)
	}
}
