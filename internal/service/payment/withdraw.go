package payment

import (
	"context"
	"errors"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/pkg/payment/processor"
	"github.com/google/uuid"
	"time"
)

var (
	ErrNotEnoughBalanceToWithdraw = errors.New("not enough balance to withdraw")
	ErrWithdrawToOwnAddress       = errors.New("withdraw to own address")
	ErrWithdrawalAmountTooSmall   = errors.New("withdrawal amount is too small")
)

const WithdrawalFee float64 = 0.001

// Withdraws amount to destination address from users wallet
// Retuns amount actually transferred after withdrawal fee or error
func WithdrawFunds(ctx context.Context, qtx *repo.Queries, userID uuid.UUID, destinationAddress string, amount int64) (int64, error) {
	fee := currency.XMR2Int(WithdrawalFee)
	if amount <= fee {
		return 0, ErrWithdrawalAmountTooSmall
	}

	wallet, err := qtx.GetWalletForUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	if wallet.BalancePico < amount {
		return 0, ErrNotEnoughBalanceToWithdraw
	}

	_, err = qtx.ReduceWalletBalance(ctx, repo.ReduceWalletBalanceParams{ID: wallet.ID, Amount: amount})
	if err != nil {
		return 0, err
	}

	withdrawAmount := amount - fee
	_, err = qtx.CreateWithdrawal(
		ctx,
		repo.CreateWithdrawalParams{
			DestinationAddress: destinationAddress,
			AmountPico:         withdrawAmount,
			Status:             repo.WithdrawalStatusPending})
	if err != nil {
		return 0, err
	}

	return withdrawAmount, nil
}

func HandleWithdrawals(ctx context.Context, qtx *repo.Queries, pp processor.Processor) error {
	if err := transferWithdrawals(ctx, qtx, pp); err != nil {
		return err
	}

	threshold := time.Now().Add(-15 * time.Minute)
	if err := handleTransactions(ctx, qtx, pp, threshold); err != nil {
		return err
	}

	return nil
}

func transferWithdrawals(ctx context.Context, qtx *repo.Queries, pp processor.Processor) error {
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

	dsts := []processor.Destination{}
	for _, w := range ws {
		dsts = append(dsts, processor.Destination{
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
func handleTransactions(ctx context.Context, qtx *repo.Queries, pp processor.Processor, threshold time.Time) error {
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
