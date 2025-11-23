package order

import (
	"context"
	"fmt"
	"github.com/gobugger/gomarket/internal/log"
	"github.com/gobugger/gomarket/internal/repo"
	"time"
)

// ProcessPaid updates orders thats invoice is paid
func ProcessPaid(ctx context.Context, qtx *repo.Queries) error {
	logger := log.Get(ctx)

	orders, err := qtx.GetOrdersByStatusXInvoiceStatus(ctx,
		repo.GetOrdersByStatusXInvoiceStatusParams{
			OrderStatus:   repo.OrderStatusPending,
			InvoiceStatus: repo.InvoiceStatusConfirmed,
		})
	if err != nil {
		return err
	}

	for _, order := range orders {
		order, err = UpdateStatus(ctx, qtx, order.ID, repo.OrderStatusPaid)
		if err != nil {
			return err
		}
		logger.Info("Order paid", "order", order)
	}
	return nil
}

// Complete dispatched orders that have been dispatched for longer than window + num_extends * window
func AutoFinalize(ctx context.Context, qtx *repo.Queries, window time.Duration) error {
	logger := log.Get(ctx)

	orders, err := qtx.GetOrdersWithStatus(ctx, repo.OrderStatusDispatched)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if shouldFinalize(&order, window) {
			if err := Finalize(ctx, qtx, order.ID); err != nil {
				return err
			}
			logger.Info("Order finalized", "orderID", order.ID)
			if err := CreateDefaultReview(ctx, qtx, order.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Cancels orders whos invoice has expired
func CancelExpired(ctx context.Context, qtx *repo.Queries) error {
	logger := log.Get(ctx)

	orders, err := qtx.GetOrdersByStatusXInvoiceStatus(ctx,
		repo.GetOrdersByStatusXInvoiceStatusParams{
			OrderStatus:   repo.OrderStatusPending,
			InvoiceStatus: repo.InvoiceStatusExpired,
		})
	if err != nil {
		return err
	}

	for _, order := range orders {
		if err := Cancel(ctx, qtx, order.ID); err != nil {
			return fmt.Errorf("failed to cancel expired order: %w", err)
		}

		logger.Info("expired order cancelled", "orderID", order.ID)
	}
	return nil
}

// Decline paid orders that are older than processingWindow
// Decline accepted orders that are accepted since dispatchWindow
func DeclineUnhandled(ctx context.Context, qtx *repo.Queries, processingWindow, dispatchWindow time.Duration) error {
	logger := log.Get(ctx)

	orders, err := qtx.GetOrdersWithStatuses(ctx, []repo.OrderStatus{
		repo.OrderStatusPaid, repo.OrderStatusAccepted,
	})
	if err != nil {
		return err
	}

	for _, order := range orders {
		if decline, reason := shouldDecline(&order, processingWindow, dispatchWindow); decline {
			if err := Decline(ctx, qtx, order.ID); err != nil {
				return err
			}
			logger.Info("order was declined because: "+reason, "orderID", order.ID)
		}
	}

	return nil
}

func shouldFinalize(order *repo.Order, deliveryWindow time.Duration) bool {
	n := time.Duration(1 + order.NumExtends)
	return time.Since(order.DispatchedAt) > deliveryWindow*n
}

func shouldDecline(order *repo.Order, processingWindow, dispatchWindow time.Duration) (should bool, reason string) {
	if order.Status == repo.OrderStatusPaid && time.Since(order.CreatedAt) > processingWindow {
		should = true
		reason = "order was not processed in time"
	} else if order.Status == repo.OrderStatusAccepted && time.Since(order.AcceptedAt) > dispatchWindow {
		should = true
		reason = "order was not dispatched in time"
	}

	return
}
