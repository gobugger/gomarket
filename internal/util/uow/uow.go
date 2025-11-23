package uow

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gobugger/gomarket/internal/repo"
)

type UoW interface {
	Do(ctx context.Context, work func(ctx context.Context, qtx *repo.Queries) error) error
	DoTx(ctx context.Context, work func(ctx context.Context, tx pgx.Tx) error) error
}

type Sql struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Sql {
	return &Sql{db: db}
}

func (uow *Sql) Do(ctx context.Context, work func(ctx context.Context, qtx *repo.Queries) error) error {
	tx, err := uow.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := work(ctx, repo.New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (uow *Sql) DoTx(ctx context.Context, work func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := uow.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := work(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
