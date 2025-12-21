package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"strings"
)

func cryptocompareURL() string {
	tsyms := ""
	supported := currency.SupportedCurrencies()
	for _, c := range supported {
		tsyms += string(c) + ","
	}
	tsyms, _ = strings.CutSuffix(tsyms, ",")
	return fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s", config.Cryptocurrency, tsyms)
}

func updateXMRPrice(ctx context.Context, db *pgxpool.Pool, c *http.Client, url string) error {
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	prices := map[currency.Currency]float64{}
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if err := currency.Set(ctx, repo.New(db), prices); err != nil {
		return err
	}

	slog.Info("prices updated", "prices", prices)

	return nil
}
