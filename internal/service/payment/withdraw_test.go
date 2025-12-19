package payment

import (
	"fmt"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/gobugger/gomarket/pkg/payment/provider/processortest"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

// Test basic withdrawal flow
func TestHandleWithdraw(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	pp := processortest.NewPaymentProvider()
	ctx := t.Context()
	qtx := repo.New(infra.Db)

	u, err := qtx.CreateUser(ctx, repo.CreateUserParams{
		Username:     "testuser",
		PasswordHash: "passwordhash",
	})
	require.NoError(t, err)

	addr, err := pp.Invoice(big.NewInt(0), "http://testhost:6969")
	require.NoError(t, err)
	pp.InvoiceStatuses[addr] = &provider.InvoiceStatus{}

	wallet, err := qtx.CreateWallet(ctx, u.ID)
	require.NoError(t, err)

	requireBalance := func(balance *big.Int) {
		w, err := qtx.GetWallet(ctx, wallet.ID)
		require.NoError(t, err)
		require.Equal(t, balance, w.BalancePico)
	}

	requireBalance(big.NewInt(0))

	_, err = qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{
		ID:     wallet.ID,
		Amount: repo.Big2Num(currency.XMR()),
	})
	require.NoError(t, err)
	requireBalance(currency.XMR())

	destAddress := testutil.XMRAddress()
	amount, err := WithdrawFunds(ctx, qtx, u.ID, destAddress, currency.XMR())
	require.NoError(t, err)
	require.Equal(t, currency.XMR().Sub(currency.XMR(), WithdrawalFee), amount)
	requireBalance(big.NewInt(0))

	ws, err := qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusPending)
	require.NoError(t, err)
	require.Equal(t, 1, len(ws))
	require.Equal(t, amount, ws[0].AmountPico)
	require.Equal(t, destAddress, ws[0].DestinationAddress)

	err = transferWithdrawals(ctx, qtx, pp)
	require.NoError(t, err)
	txs, err := qtx.GetTransactions(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(txs))

	ws, _ = qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusPending)
	require.Equal(t, 0, len(ws))
	ws, _ = qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusProcessing)
	require.Equal(t, 0, len(ws))

	pp.TransferStatuses[txs[0].Hash] = &provider.TransferStatus{}

	err = handleTransactions(ctx, qtx, pp, time.Now())
	txs, err = qtx.GetTransactions(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(txs))

	pp.TransferStatuses[txs[0].Hash] = &provider.TransferStatus{
		Confirmations: 10,
	}

	err = handleTransactions(ctx, qtx, pp, time.Now())
	txs, err = qtx.GetTransactions(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(txs))
}

// Test withdrawal flow when transfer fails
func TestHandleWithdrawFailedToTransfer(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	pp := processortest.NewPaymentProvider()
	ctx := t.Context()
	qtx := repo.New(infra.Db)

	u, err := qtx.CreateUser(ctx, repo.CreateUserParams{
		Username:     "testuser",
		PasswordHash: "passwordhash",
	})
	require.NoError(t, err)

	addr, err := pp.Invoice(big.NewInt(0), "http://testhost:6969")
	require.NoError(t, err)
	pp.InvoiceStatuses[addr] = &provider.InvoiceStatus{}

	wallet, err := qtx.CreateWallet(ctx, u.ID)
	require.NoError(t, err)

	requireBalance := func(balance *big.Int) {
		w, err := qtx.GetWallet(ctx, wallet.ID)
		require.NoError(t, err)
		require.Equal(t, balance, w.BalancePico)
	}

	requireBalance(big.NewInt(0))

	_, err = qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{
		ID:     wallet.ID,
		Amount: repo.Big2Num(currency.XMR()),
	})
	require.NoError(t, err)

	requireBalance(currency.XMR())

	withdrawAmount := currency.XMR().Rsh(currency.XMR(), 1)

	destAddress := testutil.XMRAddress()
	amount, err := WithdrawFunds(ctx, qtx, u.ID, destAddress, withdrawAmount)
	require.NoError(t, err)
	require.Equal(t, withdrawAmount.Sub(withdrawAmount, WithdrawalFee), amount)
	requireBalance(currency.XMR().Sub(currency.XMR(), withdrawAmount))

	ws, err := qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusPending)
	require.NoError(t, err)
	require.Equal(t, 1, len(ws))
	require.Equal(t, amount, ws[0].AmountPico)
	require.Equal(t, destAddress, ws[0].DestinationAddress)

	pp.TransferError = fmt.Errorf("failed to transfer withdrawal")

	err = transferWithdrawals(ctx, qtx, pp)
	require.Error(t, err)
	txs, err := qtx.GetTransactions(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(txs))

	ws, err = qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusProcessing)
	require.NoError(t, err)
	require.Equal(t, 1, len(ws))
	require.Equal(t, amount, ws[0].AmountPico)
	require.Equal(t, destAddress, ws[0].DestinationAddress)

	ws, err = qtx.GetWithdrawalsWithStatus(ctx, repo.WithdrawalStatusPending)
	require.NoError(t, err)
	require.Equal(t, 0, len(ws))
}
