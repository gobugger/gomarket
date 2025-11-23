package order

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeclineThenAccept(t *testing.T) {
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

	rfs := []float64{
		0.5, 0.8, 0, 1,
	}

	var totalQuantity int32 = 0

	for _, rf := range rfs {
		wallet, err := qtx.GetWalletForUser(ctx, customer.ID)
		require.NoError(t, err)
		initialBalance := wallet.BalancePico

		vendorWallet, err := qtx.GetWalletForUser(ctx, vendor.ID)
		require.NoError(t, err)
		initialVendorBalance := vendorWallet.BalancePico

		err = qtx.ClearCart(ctx, customer.ID)
		require.NoError(t, err)

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

		_, err = UpdateStatus(ctx, qtx, order.ID, repo.OrderStatusPaid)
		require.NoError(t, err)

		err = Accept(ctx, qtx, order.ID)
		require.NoError(t, err)
		servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusAccepted)
		totalQuantity += price.Quantity

		p, err := qtx.GetProduct(ctx, product.Product.ID)
		require.NoError(t, err)
		require.Equal(t, product.Product.Inventory-totalQuantity, p.Inventory)

		err = Dispatch(ctx, qtx, order.ID)
		require.NoError(t, err)
		servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusDispatched)

		err = Dispute(ctx, qtx, order.ID)
		require.NoError(t, err)
		servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusDisputed)

		offer, err := CreateDisputeOffer(ctx, qtx, CreateDisputeOfferParams{
			OrderID:      order.ID,
			RefundFactor: rf,
		})

		err = DeclineDisputeOffer(ctx, qtx, offer.ID)
		require.NoError(t, err)

		offer, err = CreateDisputeOffer(ctx, qtx, CreateDisputeOfferParams{
			OrderID:      order.ID,
			RefundFactor: rf,
		})

		err = AcceptDisputeOffer(ctx, qtx, offer.ID)
		require.NoError(t, err)
		servicetest.RequireOrderStatus(t, qtx, order.ID, repo.OrderStatusSettled)

		refund := order.TotalPricePico
		customerRefund := int64(rf * float64(refund))
		vendorRefund := refund - customerRefund

		wallet, err = qtx.GetWallet(ctx, wallet.ID)
		require.NoError(t, err)

		vendorWallet, err = qtx.GetWalletForUser(ctx, vendor.ID)
		require.NoError(t, err)

		require.Equal(t, initialBalance+customerRefund, wallet.BalancePico)
		require.Equal(t, initialVendorBalance+vendorRefund, vendorWallet.BalancePico)
	}
}
