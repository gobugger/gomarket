package util

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

func RiverMigrate(ctx context.Context, db *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(db), nil)
	if err != nil {
		return err
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}
