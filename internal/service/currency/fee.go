package currency

import (
	"github.com/shopspring/decimal"
)

// Fee calculates fee from amount
func Fee(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(decimal.NewFromFloat(0.05))
}

// AddFee adds fee to amount (does not modify amount)
func AddFee(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(decimal.NewFromFloat(1.05))
}

// SubFee substracts fee from amount (does not modify amount)
func SubFee(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(decimal.NewFromFloat(0.9523809523809523))
}
