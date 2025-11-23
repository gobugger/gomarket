package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"slices"
)

func validTransition(current repo.OrderStatus, next repo.OrderStatus) bool {
	return slices.Contains(getValidStatuses(next), current)
}

// Returns all valid statuses where transition to next is valid
func getValidStatuses(next repo.OrderStatus) []repo.OrderStatus {
	switch next {
	case repo.OrderStatusPending:
		return []repo.OrderStatus{}
	case repo.OrderStatusPaid, repo.OrderStatusCancelled:
		return []repo.OrderStatus{repo.OrderStatusPending}
	case repo.OrderStatusAccepted:
		return []repo.OrderStatus{repo.OrderStatusPaid}
	case repo.OrderStatusDeclined:
		return []repo.OrderStatus{repo.OrderStatusPaid, repo.OrderStatusAccepted} // vendor declined or forgot to deliver on time
	case repo.OrderStatusDispatched:
		return []repo.OrderStatus{repo.OrderStatusAccepted}
	case repo.OrderStatusFinalized, repo.OrderStatusDisputed:
		return []repo.OrderStatus{repo.OrderStatusDispatched}
	case repo.OrderStatusSettled:
		return []repo.OrderStatus{repo.OrderStatusDisputed} // Settled by admin or vendor didn't counter
	default:
		return []repo.OrderStatus{}
	}
}

func UpdateStatus(ctx context.Context, q *repo.Queries, orderID uuid.UUID, status repo.OrderStatus) (repo.Order, error) {
	order, err := q.UpdateOrderStatus(
		ctx,
		repo.UpdateOrderStatusParams{
			ID:            orderID,
			Status:        status,
			ValidStatuses: getValidStatuses(status),
		})

	if errors.Is(err, pgx.ErrNoRows) {
		return order, fmt.Errorf("can not update order status to %s: %w\n", status, ErrInvalidStatus)
	}

	return order, err
}
