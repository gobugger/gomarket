package order

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsXXX(t *testing.T) {
	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	vendor := servicetest.SetupVendor(t, infra, 0)
	customer := servicetest.SetupCustomer(t, infra, 0)
	product := servicetest.SetupProduct(t, infra, vendor.ID)

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

	require.True(t, IsVendor(ctx, qtx, IsVendorParams{UserID: vendor.ID, OrderID: order.ID}))
	require.True(t, IsCustomer(ctx, qtx, IsCustomerParams{UserID: customer.ID, OrderID: order.ID}))
}
