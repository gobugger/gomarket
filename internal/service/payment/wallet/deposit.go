package wallet

import (
	"context"
	"fmt"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"math"
)

var depositInvoiceAmount decimal.Decimal

func init() {
	if config.Cryptocurrency == "NANO" {
		var err error
		depositInvoiceAmount, err = decimal.NewFromString("100000000000000000000000000000000000000")
		if err != nil {
			panic(err)
		}
	} else {
		depositInvoiceAmount = decimal.NewFromInt(math.MaxInt64 >> 1)
	}
}

// CreateWallet creates wallet for user and an associated permanent invoices for deposits
func CreateWallet(ctx context.Context, qtx *repo.Queries, userID uuid.UUID) (repo.Wallet, error) {
	w, err := qtx.CreateWallet(ctx, userID)
	if err != nil {
		return w, err
	}

	invoice, err := qtx.CreateInvoice(ctx, repo.CreateInvoiceParams{
		AmountPico: depositInvoiceAmount,
		Permanent:  true,
	})
	if err != nil {
		return w, err
	}

	_, err = qtx.CreateDeposit(ctx, repo.CreateDepositParams{
		WalletID:  w.ID,
		InvoiceID: invoice.ID,
	})
	if err != nil {
		return w, err
	}

	return w, nil
}

func HandleDeposits(ctx context.Context, qtx *repo.Queries) error {
	deposits, err := qtx.GetOutdatedDeposits(ctx)
	if err != nil {
		return fmt.Errorf("failed to get confirmed deposits: %w", err)
	}

	logger := log.Get(ctx)

	for _, deposit := range deposits {
		unlocked, deposited := deposit.AmountUnlockedPico, deposit.AmountDepositedPico
		amount := unlocked.Sub(deposited)
		if amount.Sign() > 0 {
			_, err = qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: deposit.WalletID, Amount: amount})
			if err != nil {
				return err
			}

			_, err = qtx.UpdateAmountDeposited(ctx, repo.UpdateAmountDepositedParams{
				ID:                  deposit.ID,
				AmountDepositedPico: unlocked,
			})
			if err != nil {
				return err
			}

			logger.Info("deposit updated", "depositID", deposit.ID, "amount", amount)
		}
	}

	return nil
}
