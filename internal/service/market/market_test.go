package market

/*
import (
	"github.com/stretchr/testify/require"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/service/servicetest"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/processor/processortest"
	"testing"
)

func TestCreateOrder(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	pp := processortest.NewProcessor()
	vendorA := servicetest.SetupVendor(t, infra, 0)
	vendorB := servicetest.SetupVendor(t, infra, 0)
	customer := servicetest.SetupCustomer(t, infra, 0)
	productA := servicetest.SetupProduct(t, infra, vendorA.ID)
	productB := servicetest.SetupProduct(t, infra, vendorB.ID)

	ctx := t.Context()
	qtx := repo.New(infra.Db)
	currency.DebugStart(ctx, qtx)

	priceA := productA.Pricing[0]
	dmA := productA.DeliveryMethods[0]
	priceB := productB.Pricing[0]
	dmB := productB.DeliveryMethods[0]

	totalA, err := payment.CalculateTotalPrice(ctx, qtx,
		payment.CalculatePriceParams{
			PriceID:          priceA.ID,
			DeliveryMethodID: dmA.ID,
		})
	require.NoError(t, err)

	totalB, err := payment.CalculateTotalPrice(ctx, qtx,
		payment.CalculatePriceParams{
			PriceID:          priceB.ID,
			DeliveryMethodID: dmB.ID,
		})
	require.NoError(t, err)

	wallet, err := qtx.GetWalletForUser(ctx, customer.ID)
	require.NoError(t, err)
	initialBalance := wallet.Balance

	wallet, err = qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: wallet.ID, Amount: totalA + totalB})
	require.NoError(t, err)

	itemA, err := qtx.CreateCartItem(ctx, repo.CreateCartItemParams{
		CustomerID: customer.ID,
		PriceID:    priceA.ID,
		Quantity:   priceA.Quantity,
		Price:      priceA.Price,
	})
	require.NoError(t, err)

	itemB, err := qtx.CreateCartItem(ctx, repo.CreateCartItemParams{
		CustomerID: customer.ID,
		PriceID:    priceB.ID,
		Quantity:   priceB.Quantity,
		Price:      priceB.Price,
	})
	require.NoError(t, err)

	order, err := CreateOrder(ctx, qtx, pp, CreateOrderParams{
		CustomerID:       customer.ID,
		DeliveryMethodID: dmA.ID,
		Details:          "order details go here",
	})
	require.NoError(t, err)
	require.Equal(t, repo.OrderStatusPaid, order.Status)

	invoice, err := qtx.GetInvoiceForOrder(ctx, order.InvoiceID)
	require.NoError(t, err)
	totalXMR := currency.AddFee(currency.Fiat2XMR(currency.DefaultCurrency, itemA.Price+itemB.Price))
	require.Equal(t, totalXMR, invoice.Amount)
	servicetest.RequireBalanceForUser(t, qtx, customer.ID, initialBalance-totalXMR)

	_, err = qtx.GetOrderItems(ctx, order.ID)
	require.NoError(t, err)
}
*/
