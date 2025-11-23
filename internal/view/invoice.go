package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"time"
)

type Invoice struct {
	repo.Invoice
	TimeLeft time.Duration
}

type InvoiceView struct{}

func (iv InvoiceView) Get(ctx context.Context, q *repo.Queries, id uuid.UUID) (Invoice, error) {
	invoice, err := q.GetInvoice(ctx, id)
	if err != nil {
		return Invoice{}, err
	}

	return iv.FromModel(invoice), nil
}

func (iv InvoiceView) GetWithOrderID(ctx context.Context, q *repo.Queries, orderID uuid.UUID) (Invoice, error) {
	invoice, err := q.GetInvoiceForOrder(ctx, orderID)
	if err != nil {
		return Invoice{}, err
	}

	return iv.FromModel(invoice), nil
}

func (iv InvoiceView) FromModel(invoice repo.Invoice) Invoice {
	return Invoice{
		Invoice:  invoice,
		TimeLeft: time.Until(invoice.CreatedAt.Add(config.InvoicePaymentWindow)),
	}
}
