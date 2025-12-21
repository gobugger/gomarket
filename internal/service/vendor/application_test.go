package vendor

import (
	"crypto/rand"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/auth"
	"github.com/gobugger/gomarket/internal/service/settings"
	"github.com/gobugger/gomarket/internal/testutil"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreateApplicationAndAccept(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	ctx := t.Context()
	q := repo.New(infra.Db)

	settings, err := settings.Get(ctx, q)
	require.NoError(t, err)

	user, err := auth.Register(ctx, q, auth.RegisterParams{
		Username: "vendor" + rand.Text(),
		Password: "vendorpass123",
		PgpKey:   testutil.PgpKey,
	})
	require.NoError(t, err)

	w, err := q.CreateWallet(ctx, user.ID)
	require.NoError(t, err)

	_, err = q.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: w.ID, Amount: settings.VendorApplicationPrice})
	require.NoError(t, err)

	application, err := CreateApplication(ctx, q, infra.Mc.Client, CreateApplicationParams{
		UserID:         user.ID,
		Logo:           testutil.MockImage(t, 500, 500, "png"),
		InventoryImage: testutil.MockImage(t, 1920, 1080, "jpeg"),
		ExistingVendor: false,
		Letter:         "I'm trustworthy",
	})
	require.NoError(t, err)

	w, err = q.GetWalletForUser(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(0), w.BalancePico)
	testutil.EqualDecimal(t, settings.VendorApplicationPrice, application.PricePaidPico)

	_, err = q.GetVendorLicenseForUser(ctx, user.ID)
	require.Error(t, err)

	license, err := AcceptApplication(ctx, q, application.ID)
	require.NoError(t, err)
	testutil.EqualDecimal(t, settings.VendorApplicationPrice, license.PricePaidPico)
	require.True(t, HasLicense(ctx, q, user.ID))
}

func TestCreateApplicationExistingAndDecline(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	ctx := t.Context()
	q := repo.New(infra.Db)

	settings, err := settings.Get(ctx, q)
	require.NoError(t, err)

	user, err := auth.Register(ctx, q, auth.RegisterParams{
		Username: "vendor" + rand.Text(),
		Password: "vendorpass123",
		PgpKey:   testutil.PgpKey,
	})
	require.NoError(t, err)

	w, err := q.CreateWallet(ctx, user.ID)
	require.NoError(t, err)

	w, err = q.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: w.ID, Amount: settings.VendorApplicationPrice.Mul(decimal.NewFromInt(2))})
	require.NoError(t, err)
	initialBalance := w.BalancePico

	application, err := CreateApplication(ctx, q, infra.Mc.Client, CreateApplicationParams{
		UserID:         user.ID,
		Logo:           testutil.MockImage(t, 500, 500, "png"),
		InventoryImage: testutil.MockImage(t, 1920, 1080, "jpeg"),
		ExistingVendor: true,
		Letter:         "I'm trustworthy",
	})
	require.NoError(t, err)

	w, err = q.GetWalletForUser(ctx, user.ID)
	require.NoError(t, err)
	testutil.EqualDecimal(t, initialBalance, w.BalancePico)
	require.Equal(t, decimal.NewFromInt(0), application.PricePaidPico)

	err = DeclineApplication(ctx, q, application.ID)
	require.NoError(t, err)

	_, err = q.GetVendorApplication(ctx, application.ID)
	require.Error(t, err)
	require.False(t, HasLicense(ctx, q, user.ID))
}

func TestCreateApplicationNoBalance(t *testing.T) {
	infra := testutil.NewInfra(t.Context())
	ctx := t.Context()
	q := repo.New(infra.Db)

	settings.Set(ctx, q, settings.Settings{
		VendorApplicationPrice: decimal.NewFromInt(12018493284843), // Just random number
	})

	settings, err := settings.Get(ctx, q)
	require.NoError(t, err)

	user, err := auth.Register(ctx, q, auth.RegisterParams{
		Username: "vendor" + rand.Text(),
		Password: "vendorpass123",
		PgpKey:   testutil.PgpKey,
	})
	require.NoError(t, err)

	w, err := q.CreateWallet(ctx, user.ID)
	require.NoError(t, err)

	_, err = q.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: w.ID, Amount: settings.VendorApplicationPrice.Div(decimal.NewFromInt(2).Truncate(0))})
	require.NoError(t, err)

	_, err = CreateApplication(ctx, q, infra.Mc.Client, CreateApplicationParams{
		UserID:         user.ID,
		Logo:           testutil.MockImage(t, 500, 500, "png"),
		InventoryImage: testutil.MockImage(t, 1920, 1080, "jpeg"),
		ExistingVendor: false,
		Letter:         "I'm trustworthy",
	})
	require.ErrorIs(t, err, ErrNotEnoughBalance)
	require.False(t, HasLicense(ctx, q, user.ID))
}
