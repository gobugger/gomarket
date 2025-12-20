package repo

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestNumericConversion(t *testing.T) {
	nums := []int64{
		27498327498247,
		0,
		-23749274397483,
	}

	for _, n := range nums {
		nb := big.NewInt(n)
		nc := Big2Num(nb)
		require.Equal(t, nb, Num2Big(nc))
	}
}
