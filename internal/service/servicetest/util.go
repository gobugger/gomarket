package servicetest

import (
	"crypto/rand"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/auth"
	"github.com/gobugger/gomarket/internal/service/payment/wallet"
	"github.com/gobugger/gomarket/internal/service/product"
	"github.com/gobugger/gomarket/internal/service/settings"
	license "github.com/gobugger/gomarket/internal/service/vendor"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/gobugger/gomarket/pkg/payment/provider/processortest"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

// Creates vendor with random name and zero balance
func SetupVendor(t *testing.T, infra *testutil.Infra, balance decimal.Decimal) *repo.User {
	ctx := t.Context()
	q := repo.New(infra.Db)

	err := settings.Set(ctx, q, settings.Settings{VendorApplicationPrice: decimal.NewFromInt(1000000000000)})
	require.NoError(t, err)
	settings, err := settings.Get(ctx, q)
	require.NoError(t, err)

	vendor, err := auth.Register(ctx, q, auth.RegisterParams{Username: rand.Text()[0:13], Password: "supersecret123!", PgpKey: testutil.PgpKey})
	require.NoError(t, err)

	vw, err := wallet.CreateWallet(ctx, q, vendor.ID)
	require.NoError(t, err)

	_, err = q.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: vw.ID, Amount: settings.VendorApplicationPrice.Add(balance)})
	require.NoError(t, err)

	va, err := license.CreateApplication(ctx, q, infra.Mc.Client, license.CreateApplicationParams{
		UserID:         vendor.ID,
		Logo:           testutil.MockImage(t, 500, 500, "png"),
		InventoryImage: testutil.MockImage(t, 1920, 1080, "jpeg"),
		ExistingVendor: false,
		Letter:         "",
	})
	require.NoError(t, err)

	vw, err = q.GetWallet(ctx, vw.ID)
	require.NoError(t, err)
	require.Equal(t, balance, vw.BalancePico)

	l, err := license.AcceptApplication(ctx, q, va.ID)
	require.NoError(t, err)
	require.Equal(t, settings.VendorApplicationPrice, l.PricePaidPico)

	_, err = q.CreateDeliveryMethod(ctx, repo.CreateDeliveryMethodParams{
		Description: "post delivery",
		PriceCent:   500,
		VendorID:    vendor.ID,
	})
	require.NoError(t, err)

	return &vendor
}

// Creates basic account with balance
func SetupCustomer(t *testing.T, infra *testutil.Infra, balance decimal.Decimal) *repo.User {
	ctx := t.Context()

	q := repo.New(infra.Db)

	customer, err := auth.Register(ctx, q, auth.RegisterParams{
		Username: rand.Text()[0:13],
		Password: "userpass123",
		PgpKey:   "",
	})
	require.NoError(t, err)

	cw, err := wallet.CreateWallet(ctx, q, customer.ID)
	require.NoError(t, err)

	_, err = q.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: cw.ID, Amount: balance})
	require.NoError(t, err)

	return &customer
}

type TestProduct struct {
	Product *repo.Product
	Pricing []repo.PriceTier
}

// Creates product listing for vendor
func SetupProduct(t *testing.T, infra *testutil.Infra, vendorID uuid.UUID) *TestProduct {
	ctx := t.Context()
	q := repo.New(infra.Db)

	categories, err := q.GetCategories(ctx)
	require.NoError(t, err)
	if len(categories) == 0 {
		category, err := q.CreateCategory(ctx, repo.CreateCategoryParams{
			Name: "berry",
		})
		require.NoError(t, err)
		categories = append(categories, category)
	}

	p, err := product.Create(
		ctx,
		q,
		infra.Mc.Client,
		product.CreateParams{
			Title:       "Product title",
			Description: "Product description",
			Image:       testutil.MockImage(t, 1920, 1080, "jpeg"),
			CategoryID:  categories[0].ID,
			PriceTiers:  []product.PriceTier{{Quantity: 1, PriceCent: 2000}, {Quantity: 2, PriceCent: 3000}},
			Inventory:   1000,
			ShipsFrom:   product.SupportedLocations()[0],
			ShipsTo:     product.SupportedLocations()[0],
			VendorID:    vendorID,
		})
	require.NoError(t, err)

	prices, err := q.GetPriceTiers(ctx, p.ID)
	require.NoError(t, err)

	return &TestProduct{&p, prices}
}

func CreateInvoice(t *testing.T, q *repo.Queries, amount decimal.Decimal) repo.Invoice {
	ctx := t.Context()
	invoice, err := q.CreateInvoice(ctx, repo.CreateInvoiceParams{
		AmountPico: amount,
	})
	require.NoError(t, err)
	_, err = q.SetInvoiceAddress(ctx, repo.SetInvoiceAddressParams{
		ID:      invoice.ID,
		Address: testutil.XMRAddress(),
	})
	require.NoError(t, err)

	return invoice
}

func PayInvoice(pp *processortest.PaymentProvider, invoice *repo.Invoice) {
	pp.InvoiceStatuses[invoice.Address] = &provider.InvoiceStatus{
		AmountUnlocked: invoice.AmountPico,
		AmountTotal:    invoice.AmountPico,
	}
}

func RequireOrderStatus(t *testing.T, q *repo.Queries, orderID uuid.UUID, status repo.OrderStatus) {
	o, err := q.GetOrder(t.Context(), orderID)
	require.NoError(t, err)
	require.Equal(t, status, o.Status)
}

func RequireBalanceForUser(t *testing.T, q *repo.Queries, userID uuid.UUID, balance decimal.Decimal) {
	cw, err := q.GetWalletForUser(t.Context(), userID)
	require.NoError(t, err)
	testutil.EqualDecimal(t, balance, cw.BalancePico)
}
