package repo

import (
	"github.com/jackc/pgx/v5/pgtype"
	"math/big"
)

func Num2Big(num pgtype.Numeric) *big.Int {
	if !num.Valid || num.Exp != 0 {
		return nil
	}

	return new(big.Int).Set(num.Int)
}

func Big2Num(num *big.Int) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   num,
		Exp:   0,
		Valid: num != nil,
	}
}
