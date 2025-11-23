package db

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func ErrCode(err error) string {
	var pgErr *pgconn.PgError
	if err == nil || !errors.As(err, &pgErr) {
		return ""
	}

	return pgErr.Code
}
