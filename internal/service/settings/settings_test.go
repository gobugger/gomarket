package settings

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestSettings(t *testing.T) {
	db, cleanup, dsn := testutil.SetupDB()
	defer cleanup()

	q := repo.New(db)
	settings, err := Get(t.Context(), q)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1000000000000), settings.VendorApplicationPrice)

	err = Set(t.Context(), q, Settings{VendorApplicationPrice: big.NewInt(420)})
	require.NoError(t, err)

	settings, err = Get(t.Context(), q)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(420), settings.VendorApplicationPrice)

	testutil.CleanDB(dsn)

	settings, err = Get(t.Context(), q)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1000000000000), settings.VendorApplicationPrice)
}
