package order

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestShouldFinalize(t *testing.T) {
	dt := 4 * time.Millisecond

	order := repo.Order{
		DispatchedAt: time.Now(),
	}

	require.False(t, shouldFinalize(&order, dt))
	time.Sleep(dt)
	require.True(t, shouldFinalize(&order, dt))

	order.NumExtends++
	require.False(t, shouldFinalize(&order, dt))
	time.Sleep(dt)
	require.True(t, shouldFinalize(&order, dt))

	order.NumExtends++
	require.False(t, shouldFinalize(&order, dt))
	time.Sleep(dt)
	require.True(t, shouldFinalize(&order, dt))
}

func TestShouldDecline(t *testing.T) {
	dt := 4 * time.Millisecond

	order := repo.Order{
		Status:    repo.OrderStatusPaid,
		CreatedAt: time.Now(),
	}

	decline, _ := shouldDecline(&order, dt, dt)
	require.False(t, decline)
	time.Sleep(dt)
	decline, reason := shouldDecline(&order, dt, dt)
	require.True(t, decline)
	require.Contains(t, reason, "processed")

	order = repo.Order{
		Status:     repo.OrderStatusAccepted,
		AcceptedAt: time.Now(),
	}

	decline, _ = shouldDecline(&order, dt, dt)
	require.False(t, decline)
	time.Sleep(dt)
	decline, reason = shouldDecline(&order, dt, dt)
	require.True(t, decline)
	require.Contains(t, reason, "dispatched")
}

func TestProcessPaidAndAutoFinalize(t *testing.T) {
	vendor := servicetest.SetupVendor(t, infra, decimal.NewFromInt(0))
	customer := servicetest.SetupCustomer(t, infra, decimal.NewFromInt(0))
	product := servicetest.SetupProduct(t, infra, vendor.ID)

	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	price := product.Pricing[0]
	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, vendor.ID)
	require.NoError(t, err)
	dm := dms[0]

	_, err = qtx.CreateCartItem(ctx, repo.CreateCartItemParams{
		CustomerID: customer.ID,
		PriceID:    price.ID,
	})
	require.NoError(t, err)

	order, err := Create(ctx, qtx, CreateParams{
		DeliveryMethodID: dm.ID,
		CustomerID:       customer.ID,
		Details:          "Order details here",
	})
	require.NoError(t, err)
	require.Equal(t, repo.OrderStatusPending, order.Status)

	err = ProcessPaid(ctx, qtx)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusPending)

	invoice := servicetest.CreateInvoice(t, qtx, currency.AddFee(order.TotalPricePico))
	_, err = qtx.CreateOrderInvoice(ctx, repo.CreateOrderInvoiceParams{
		OrderID:   order.ID,
		InvoiceID: invoice.ID,
	})
	require.NoError(t, err)

	err = ProcessPaid(ctx, qtx)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusPending)

	_, err = qtx.UpdateInvoiceStatus(ctx, repo.UpdateInvoiceStatusParams{
		ID:     invoice.ID,
		Status: repo.InvoiceStatusConfirmed,
	})
	require.NoError(t, err)

	err = ProcessPaid(ctx, qtx)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusPaid)

	err = Accept(ctx, qtx, order.ID)
	require.NoError(t, err)

	err = Dispatch(ctx, qtx, order.ID)
	require.NoError(t, err)

	err = AutoFinalize(ctx, qtx, time.Minute)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusDispatched)

	time.Sleep(time.Millisecond)

	err = AutoFinalize(ctx, qtx, time.Millisecond)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusFinalized)
}

func TestCancelExpired(t *testing.T) {
	vendor := servicetest.SetupVendor(t, infra, decimal.NewFromInt(0))
	customer := servicetest.SetupCustomer(t, infra, decimal.NewFromInt(0))
	product := servicetest.SetupProduct(t, infra, vendor.ID)

	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	price := product.Pricing[0]
	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, vendor.ID)
	require.NoError(t, err)
	dm := dms[0]

	_, err = qtx.CreateCartItem(ctx, repo.CreateCartItemParams{
		CustomerID: customer.ID,
		PriceID:    price.ID,
	})
	require.NoError(t, err)

	order, err := Create(ctx, qtx, CreateParams{
		DeliveryMethodID: dm.ID,
		CustomerID:       customer.ID,
		Details:          "Order details here",
	})
	require.NoError(t, err)
	require.Equal(t, repo.OrderStatusPending, order.Status)

	invoice := servicetest.CreateInvoice(t, qtx, currency.AddFee(order.TotalPricePico))
	_, err = qtx.CreateOrderInvoice(ctx, repo.CreateOrderInvoiceParams{
		OrderID:   order.ID,
		InvoiceID: invoice.ID,
	})
	require.NoError(t, err)

	err = CancelExpired(ctx, qtx)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusPending)

	_, err = qtx.UpdateInvoiceStatus(ctx, repo.UpdateInvoiceStatusParams{
		ID:     invoice.ID,
		Status: repo.InvoiceStatusExpired,
	})
	require.NoError(t, err)
	err = CancelExpired(ctx, qtx)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusCancelled)
}
