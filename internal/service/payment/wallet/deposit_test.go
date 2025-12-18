package wallet

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/gobugger/gomarket/pkg/payment/provider/processortest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func requireBalance(t *testing.T, qtx *repo.Queries, balance int64, walletID uuid.UUID) {
	wallet, err := qtx.GetWallet(t.Context(), walletID)
	require.NoError(t, err)
	require.Equal(t, balance, wallet.BalancePico)
}

func TestHandleDeposits(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	pp := processortest.NewPaymentProvider()

	ctx := t.Context()
	qtx := repo.New(infra.Db)

	user, err := qtx.CreateUser(ctx, repo.CreateUserParams{
		Username:     "someuser",
		PasswordHash: "passwordhash",
	})
	require.NoError(t, err)

	wallet, err := CreateWallet(ctx, qtx, user.ID)
	require.NoError(t, err)
	initialBalance := wallet.BalancePico

	deposit, err := qtx.GetDepositForWallet(ctx, wallet.ID)
	require.NoError(t, err)

	err = payment.PrepareInvoices(ctx, qtx, pp)
	require.NoError(t, err)

	deposit, err = qtx.GetDepositForWallet(ctx, wallet.ID)
	require.NoError(t, err)

	deposits := []int64{1, 123456789, 1e12 / 2, 1e12, 10 * 1e12, 100 * 1e12, 1000 * 1e12}
	total := initialBalance

	for _, d := range deposits {
		total += d
		pp.InvoiceStatuses[deposit.Invoice.Address] = &provider.InvoiceStatus{
			AmountUnlocked: total,
		}

		err = payment.ProcessInvoices(ctx, qtx, pp, time.Hour)
		require.NoError(t, err)

		for range 2 {
			err = HandleDeposits(ctx, qtx)
			require.NoError(t, err)
		}

		requireBalance(t, qtx, total, wallet.ID)
	}
}
