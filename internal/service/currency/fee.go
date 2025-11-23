package currency

// Fee calculates fee from amount
func Fee(amount int64) int64 {
	return int64(float64(amount) * 0.05)
}

// AddFee adds fee to amount
func AddFee(amount int64) int64 {
	return int64(float64(amount) * 1.05)
}

// SubFee substracts fee from amount
func SubFee(amount int64) int64 {
	return int64(float64(amount) * 0.9523809523809523)
}
