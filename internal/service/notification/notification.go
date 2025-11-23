package notification

import (
	"context"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/localizer"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

func NewOrder(ctx context.Context, q *repo.Queries, orderID uuid.UUID) {
	vendor, err := q.GetVendorForOrder(ctx, orderID)
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
		return
	}

	l, _ := localizer.Get(vendor.Locale)

	_, err = q.CreateNotification(ctx,
		repo.CreateNotificationParams{
			Title: l.Translate("Order received"),
			Content: l.Translate(`You have received a new order!
You have a total of %d hours to accept, before it's auto-declined.`, config.OrderProcessingWindow/(time.Hour)),
			UserID: vendor.ID,
		})
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
	}
}

func OrderAccepted(ctx context.Context, q *repo.Queries, orderID uuid.UUID) {
	customer, err := q.GetCustomerForOrder(ctx, orderID)
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
		return
	}

	l, _ := localizer.Get(customer.Locale)

	_, err = q.CreateNotification(ctx,
		repo.CreateNotificationParams{
			Title:   l.Translate("Order accepted"),
			Content: l.Translate("Your order has been accepted!\nYou can monitor its status from orders page"),
			UserID:  customer.ID,
		})
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
	}
}

func OrderDispatched(ctx context.Context, q *repo.Queries, orderID uuid.UUID) {
	customer, err := q.GetCustomerForOrder(ctx, orderID)
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
		return
	}

	l, _ := localizer.Get(customer.Locale)

	_, err = q.CreateNotification(ctx,
		repo.CreateNotificationParams{
			Title:   l.Translate("Order dispatched"),
			Content: l.Translate("Your order has been dispatched!\nPlease remember to review it."),
			UserID:  customer.ID,
		})
	if err != nil {
		slog.Error("notification", slog.Any("error", err))
	}
}
