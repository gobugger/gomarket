package testutil

import (
	"crypto/rand"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func EqualDecimal(t *testing.T, a, b decimal.Decimal) {
	require.True(t, a.Equal(b), fmt.Sprintf("decimal %v does not equal %v", a, b))
}

func XMRAddress() string {
	addr := rand.Text()
	for len(addr) < 95 {
		addr += rand.Text()
	}
	return addr[:95]
}

func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for range 10 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("project root not found from %s", dir)
}

const PgpKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
GKxRQDlxtgfojbYLd7F6W8jhuMqH
=QDn9
-----END PGP PUBLIC KEY BLOCK-----`
