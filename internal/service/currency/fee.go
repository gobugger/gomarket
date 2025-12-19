package currency

import (
	"math/big"
)

// Fee calculates fee from amount
func Fee(amount *big.Int) *big.Int {
	return mul(amount, big.NewFloat(0.05))
}

// AddFee adds fee to amount (does not modify amount)
func AddFee(amount *big.Int) *big.Int {
	return mul(amount, big.NewFloat(1.05))
}

// SubFee substracts fee from amount (does not modify amount)
func SubFee(amount *big.Int) *big.Int {
	return mul(amount, big.NewFloat(0.9523809523809523))
}

func mul(amount *big.Int, x *big.Float) *big.Int {
	af := new(big.Float).SetInt(amount)
	result, _ := af.Mul(af, x).Int(nil)
	return result
}
