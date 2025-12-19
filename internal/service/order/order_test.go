package order

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/gobugger/gomarket/internal/testutil"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var infra *testutil.Infra

func TestMain(m *testing.M) {
	infra = testutil.NewInfra(context.Background())

	m.Run()

	infra.Close()
}

func TestFinalize(t *testing.T) {
	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	vendor := servicetest.SetupVendor(t, infra, big.NewInt(0))
	customer := servicetest.SetupCustomer(t, infra, big.NewInt(0))
	product := servicetest.SetupProduct(t, infra, vendor.ID)

	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, vendor.ID)
	require.NoError(t, err)
	dm := dms[0]
	price := product.Pricing[0]

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

	err = Accept(ctx, qtx, order.ID)
	require.Error(t, err)

	_, err = UpdateStatus(ctx, qtx, order.ID, repo.OrderStatusPaid)
	require.NoError(t, err)

	err = Accept(ctx, qtx, order.ID)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusAccepted)

	p, err := qtx.GetProduct(ctx, product.Product.ID)
	require.NoError(t, err)
	require.Equal(t, product.Product.Inventory-price.Quantity, p.Inventory)

	err = Dispatch(ctx, qtx, order.ID)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusDispatched)

	err = Finalize(ctx, qtx, order.ID)
	require.NoError(t, err)
	servicetest.RequireBalanceForUser(t, qtx, vendor.ID, repo.Num2Big(order.TotalPricePico))
}

func TestCancel(t *testing.T) {
	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	vendor := servicetest.SetupVendor(t, infra, big.NewInt(0))
	customer := servicetest.SetupCustomer(t, infra, big.NewInt(0))
	product := servicetest.SetupProduct(t, infra, vendor.ID)

	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, vendor.ID)
	require.NoError(t, err)
	dm := dms[0]
	price := product.Pricing[0]

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

	err = Cancel(ctx, qtx, order.ID)
	require.NoError(t, err)

	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusCancelled)
	servicetest.RequireBalanceForUser(t, qtx, vendor.ID, big.NewInt(0))
}

func TestDecline(t *testing.T) {
	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	vendor := servicetest.SetupVendor(t, infra, big.NewInt(0))
	customer := servicetest.SetupCustomer(t, infra, big.NewInt(0))
	product := servicetest.SetupProduct(t, infra, vendor.ID)

	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, vendor.ID)
	require.NoError(t, err)
	dm := dms[0]
	price := product.Pricing[0]

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

	err = Decline(ctx, qtx, order.ID)
	require.Error(t, err)

	_, err = UpdateStatus(ctx, qtx, order.ID, repo.OrderStatusPaid)
	require.NoError(t, err)

	err = Decline(ctx, qtx, order.ID)
	require.NoError(t, err)
	servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusDeclined)

	servicetest.RequireBalanceForUser(t, qtx, customer.ID, currency.AddFee(repo.Num2Big(order.TotalPricePico)))
	servicetest.RequireBalanceForUser(t, qtx, vendor.ID, big.NewInt(0))
}
