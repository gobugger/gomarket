package order

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
)

type IsCustomerParams struct {
	UserID  uuid.UUID
	OrderID uuid.UUID
}

func IsCustomer(ctx context.Context, qtx *repo.Queries, p IsCustomerParams) bool {
	order, err := qtx.GetOrder(ctx, p.OrderID)
	return err == nil && p.UserID == order.CustomerID
}

type IsVendorParams struct {
	UserID  uuid.UUID
	OrderID uuid.UUID
}

func IsVendor(ctx context.Context, qtx *repo.Queries, p IsVendorParams) bool {
	vendor, err := qtx.GetVendorForOrder(ctx, p.OrderID)
	return err == nil && p.UserID == vendor.ID
}

type IsCustomerOrVendorParams struct {
	UserID  uuid.UUID
	OrderID uuid.UUID
}

func IsCustomerOrVendor(ctx context.Context, qtx *repo.Queries, p IsCustomerOrVendorParams) bool {
	order, err := qtx.GetOrder(ctx, p.OrderID)
	return err == nil && (p.UserID == order.CustomerID || p.UserID == order.VendorID)
}
