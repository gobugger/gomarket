package rand

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntn(t *testing.T) {
	m := 100
	n := 100 * m
	var total int64
	for range n {
		total += int64(Intn(m))
	}

	avg := float64(total) / float64(n)

	require.InDelta(t, avg, (m-1)/2, 1)
}
