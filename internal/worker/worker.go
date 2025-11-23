package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/pkg/payment/processor"
)

type PrepareInvoiceArgs struct {
	ID uuid.UUID
}

func (PrepareInvoiceArgs) Kind() string { return "prepare_invoice" }

type PrepareInvoiceWorker struct {
	Db *pgxpool.Pool
	Pp processor.Processor
	river.WorkerDefaults[PrepareInvoiceArgs]
}

func (w *PrepareInvoiceWorker) Work(ctx context.Context, job *river.Job[PrepareInvoiceArgs]) error {
	tx, err := w.Db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := repo.New(tx)
	if err = payment.PrepareInvoice(ctx, qtx, w.Pp, job.Args.ID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
