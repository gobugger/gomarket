package payment

import (
	"context"
	"fmt"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/pkg/payment/provider"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

// CreateInvoice creates a penging, expiring invoice
func CreateInvoice(ctx context.Context, qtx *repo.Queries, amount decimal.Decimal) (repo.Invoice, error) {
	// TODO: Trigger event to prepare invoice
	return qtx.CreateInvoice(ctx, repo.CreateInvoiceParams{
		AmountPico: amount,
		Permanent:  false,
	})
}

// Assignes address to all pending invoices that don't have an address yet
func PrepareInvoice(ctx context.Context, qtx *repo.Queries, pp provider.PaymentProvider, id uuid.UUID) error {
	invoice, err := qtx.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	if invoice.Address != "" {
		return fmt.Errorf("invoice %s already has an address %s", id, invoice.Address)
	}

	address, err := pp.Invoice(invoice.AmountPico, "")
	if err != nil {
		return fmt.Errorf("failed to generate invoice address: %w", err)
	}

	_, err = qtx.SetInvoiceAddress(ctx, repo.SetInvoiceAddressParams{
		ID:      invoice.ID,
		Address: address,
	})

	return err
}

func ProcessInvoices(ctx context.Context, qtx *repo.Queries, pp provider.PaymentProvider, paymentWindow time.Duration) error {
	invoices, err := qtx.GetPendingInvoices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending invoices: %w", err)
	}

	logger := log.Get(ctx)

	for _, invoice := range invoices {
		status, err := pp.InvoiceStatus(invoice.Address)
		if err != nil {
			return err
		}

		amountUnlocked, amount := invoice.AmountUnlockedPico, invoice.AmountPico

		if status.AmountUnlocked.Cmp(amountUnlocked) > 0 {
			_, err := qtx.UpdateInvoiceAmountUnlocked(ctx, repo.UpdateInvoiceAmountUnlockedParams{
				ID:                 invoice.ID,
				AmountUnlockedPico: status.AmountUnlocked,
			})
			if err != nil {
				return fmt.Errorf("failed to update invoice amount unlocked: %w", err)
			}
		}

		if status.AmountUnlocked.Cmp(amount) >= 0 {
			_, err := qtx.UpdateInvoiceStatus(ctx, repo.UpdateInvoiceStatusParams{
				ID:     invoice.ID,
				Status: repo.InvoiceStatusConfirmed,
			})
			if err != nil {
				return fmt.Errorf("failed to update invoice status: %w", err)
			}

			logger.Info("invoice confirmed", "invoiceID", invoice.ID, "amount", invoice.AmountPico)
		} else if !invoice.Permanent && time.Since(invoice.CreatedAt) > paymentWindow {
			_, err := qtx.UpdateInvoiceStatus(ctx, repo.UpdateInvoiceStatusParams{
				ID:     invoice.ID,
				Status: repo.InvoiceStatusExpired,
			})
			if err != nil {
				return fmt.Errorf("failed to update invoice status: %w", err)
			}

			logger.Info("invoice expired", "invoiceID", invoice.ID)
		}
	}

	return nil
}
