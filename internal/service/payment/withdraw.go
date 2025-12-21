package payment

import (
	"context"
	"errors"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

var (
	ErrNotEnoughBalanceToWithdraw = errors.New("not enough balance to withdraw")
	ErrWithdrawToOwnAddress       = errors.New("withdraw to own address")
	ErrWithdrawalAmountTooSmall   = errors.New("withdrawal amount is too small")
)

var WithdrawalFee decimal.Decimal

func init() {
	if config.Cryptocurrency == "NANO" {
		var err error
		WithdrawalFee, err = decimal.NewFromString("10000000000000000000000000000")
		if err != nil {
			panic(err)
		}
	} else {
		WithdrawalFee = decimal.NewFromInt(100000000)
	}
}

// Withdraws amount to destination address from users wallet
// Retuns amount actually transferred after withdrawal fee or error
func WithdrawFunds(ctx context.Context, qtx *repo.Queries, userID uuid.UUID, destinationAddress string, amount decimal.Decimal) (decimal.Decimal, error) {
	if amount.Cmp(WithdrawalFee) <= 0 {
		return decimal.NewFromInt(0), ErrWithdrawalAmountTooSmall
	}

	wallet, err := qtx.GetWalletForUser(ctx, userID)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	if wallet.BalancePico.Cmp(amount) < 0 {
		return decimal.NewFromInt(0), ErrNotEnoughBalanceToWithdraw
	}

	_, err = qtx.ReduceWalletBalance(ctx, repo.ReduceWalletBalanceParams{ID: wallet.ID, Amount: amount})
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	withdrawAmount := amount.Sub(WithdrawalFee)
	_, err = qtx.CreateWithdrawal(
		ctx,
		repo.CreateWithdrawalParams{
			DestinationAddress: destinationAddress,
			AmountPico:         withdrawAmount,
			Status:             repo.WithdrawalStatusPending})
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	return withdrawAmount, nil
}

func HandleWithdrawals(ctx context.Context, qtx *repo.Queries, pp provider.PaymentProvider) error {
	if err := transferWithdrawals(ctx, qtx, pp); err != nil {
		return err
	}

	threshold := time.Now().Add(-15 * time.Minute)
	if err := handleTransactions(ctx, qtx, pp, threshold); err != nil {
		return err
	}

	return nil
}

func transferWithdrawals(ctx context.Context, qtx *repo.Queries, pp provider.PaymentProvider) error {
	ws, err := qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusPending)
	if err != nil {
		return err
	}
	if len(ws) == 0 {
		return nil
	}

	logger := log.Get(ctx)

	for _, w := range ws {
		_, err := qtx.UpdateWithdrawalStatus(
			ctx,
			repo.UpdateWithdrawalStatusParams{
				ID:     w.ID,
				Status: repo.WithdrawalStatusProcessing,
			})
		if err != nil {
			return err
		}
	}

	dsts := []provider.Destination{}
	for _, w := range ws {
		dsts = append(dsts, provider.Destination{
			Amount:  w.AmountPico,
			Address: w.DestinationAddress,
		})
	}

	resp, err := pp.Transfer(dsts)
	if err != nil {
		logger.Error("Transfer failed", "destinations", dsts)
		return err
	}

	logger.Info("Initiated transactions", "destinations", dsts, "txHashList", resp.TxHashList)

	onError := func() {
		logger.Error("Failed to delete withdrawals and create transactions", "withdrawals", ws, "transactions", resp.TxHashList)
	}

	for _, w := range ws {
		if err := qtx.DeleteWithdrawal(ctx, w.ID); err != nil {
			logger.Error("failed to delete withdrawal", "withdrawal", w)
			onError()
			return err
		}
	}

	for _, txHash := range resp.TxHashList {
		if _, err := qtx.CreateTransaction(ctx, txHash); err != nil {
			logger.Error("failed to create transaction", "txHash", txHash)
			onError()
			return err
		}
	}

	return nil
}

// handleTransactions deletes confirmed and logs failed transactions
func handleTransactions(ctx context.Context, qtx *repo.Queries, pp provider.PaymentProvider, threshold time.Time) error {
	txs, err := qtx.GetTransactionsBefore(ctx, threshold)
	if err != nil {
		return err
	}

	logger := log.Get(ctx)

	for _, tx := range txs {
		status, err := pp.TransferStatus(tx.Hash)
		if err != nil {
			return err
		}

		if status.Confirmations >= 10 {
			if err := qtx.DeleteTransaction(ctx, tx.ID); err != nil {
				return err
			}
			logger.Info("transaction confirmed and deleted", "txHash", tx.Hash)
		} else if status.Failed {
			logger.Error("transaction has failed", "txHash", tx.Hash)
		}
	}

	return nil
}
