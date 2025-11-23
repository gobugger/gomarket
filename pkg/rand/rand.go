package rand

import (
	"crypto/rand"
	"math/big"
)

// Return int in range [0, max).
func Intn(max int) int {
	if max <= 0 {
		return 0
	}

	val, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(val.Int64())
}

// Returns random float64 in range [0, 1)
func Float64() float64 {
	n := 1000000
	return float64(Intn(n)) / float64(n)
}
