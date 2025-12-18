package payment

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/gobugger/gomarket/pkg/payment/provider/processortest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func requireInvoiceStatus(t *testing.T, qtx *repo.Queries, status repo.InvoiceStatus, invoiceID uuid.UUID) {
	invoice, err := qtx.GetInvoice(t.Context(), invoiceID)
	require.NoError(t, err)
	require.Equal(t, status, invoice.Status)
}

func TestInvoicing(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	pp := processortest.NewPaymentProvider()

	ctx := t.Context()
	qtx := repo.New(infra.Db)

	user, err := qtx.CreateUser(ctx, repo.CreateUserParams{
		Username:     "someuser",
		PasswordHash: "passwordhash",
	})
	require.NoError(t, err)

	_, err = qtx.CreateWallet(ctx, user.ID)
	require.NoError(t, err)

	amount := currency.XMR

	invoice := servicetest.CreateInvoice(t, qtx, amount)
	require.Equal(t, repo.InvoiceStatusPending, invoice.Status)

	err = ProcessInvoices(ctx, qtx, pp, time.Minute)
	require.NoError(t, err)
	invoice, err = qtx.GetInvoice(ctx, invoice.ID)
	require.NoError(t, err)
	require.Equal(t, repo.InvoiceStatusPending, invoice.Status)

	pp.InvoiceStatuses[invoice.Address] = &provider.InvoiceStatus{
		AmountUnlocked: amount,
		AmountTotal:    amount,
	}

	err = ProcessInvoices(ctx, qtx, pp, time.Minute)
	require.NoError(t, err)
	invoice, err = qtx.GetInvoice(ctx, invoice.ID)
	require.NoError(t, err)
	require.Equal(t, repo.InvoiceStatusConfirmed, invoice.Status)

	invoice = servicetest.CreateInvoice(t, qtx, amount)
	invoice, err = qtx.GetInvoice(ctx, invoice.ID)
	require.NoError(t, err)
	require.Equal(t, repo.InvoiceStatusPending, invoice.Status)

	time.Sleep(time.Millisecond)

	err = ProcessInvoices(ctx, qtx, pp, time.Millisecond)
	require.NoError(t, err)
	invoice, err = qtx.GetInvoice(ctx, invoice.ID)
	require.NoError(t, err)
	require.Equal(t, repo.InvoiceStatusExpired, invoice.Status)
}
