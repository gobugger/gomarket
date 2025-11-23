package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/currency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

var (
	ErrNotEnoughBalance          = errors.New("not enough balance")
	ErrCustomerIsVendor          = errors.New("customer can't be vendor")
	ErrInvalidStatus             = errors.New("invalid status")
	ErrExtendUnavailable         = errors.New("extend unavailable")
	ErrReExtendUnavailable       = errors.New("re-extend unavailable")
	ErrUnableToExtendFurther     = errors.New("unable to extend further")
	ErrOrderIsEmpty              = errors.New("order is empty")
	ErrPricingUnavailable        = errors.New("pricing unavailable")
	ErrDeliveryMethodUnavailable = errors.New("delivery method unavailable")
	ErrCartIsEmpty               = errors.New("cart is empty")
)

type CreateParams struct {
	CustomerID       uuid.UUID
	DeliveryMethodID uuid.UUID
	Details          string
}

// Create an order based on items in the customers cart
func Create(ctx context.Context, qtx *repo.Queries, p CreateParams) (repo.Order, error) {
	dm, err := qtx.GetDeliveryMethod(ctx, p.DeliveryMethodID)
	if err != nil {
		return repo.Order{}, err
	} else if dm.DeletedAt.Valid {
		return repo.Order{}, ErrDeliveryMethodUnavailable
	}

	if p.CustomerID == dm.VendorID {
		return repo.Order{}, ErrCustomerIsVendor
	}

	priceTiers, err := qtx.GetPriceTiersForCart(ctx, repo.GetPriceTiersForCartParams{
		CustomerID: p.CustomerID,
		VendorID:   dm.VendorID,
	})
	if err != nil {
		return repo.Order{}, err
	}

	if len(priceTiers) == 0 {
		return repo.Order{}, ErrCartIsEmpty
	}

	total := dm.PriceCent
	for _, pt := range priceTiers {
		if pt.DeletedAt.Valid {
			return repo.Order{}, ErrPricingUnavailable
		}

		total += pt.PriceCent * pt.Count
	}

	totalXMR := currency.Fiat2XMR(currency.DefaultCurrency, total)

	order, err := qtx.CreateOrder(
		ctx,
		repo.CreateOrderParams{
			DeliveryMethodID: p.DeliveryMethodID,
			CustomerID:       p.CustomerID,
			TotalPricePico:   totalXMR,
			Details:          p.Details,
		})
	if err != nil {
		return repo.Order{}, fmt.Errorf("failed to create order: %w\n", err)
	}

	itemsParams := []repo.CreateOrderItemsParams{}
	for _, pt := range priceTiers {
		itemsParams = append(itemsParams, repo.CreateOrderItemsParams{
			OrderID: order.ID,
			PriceID: pt.ID,
			Count:   int32(pt.Count),
		})
	}

	if _, err = qtx.CreateOrderItems(ctx, itemsParams); err != nil {
		return repo.Order{}, err
	}

	return order, nil
}

// Cancel updates order status to cancelled
func Cancel(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusCancelled)
	return err
}

// Accept sets order status to accepted and updates inventory
func Accept(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusAccepted)
	if err != nil {
		return err
	}

	items, err := qtx.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}

	productAmounts := map[uuid.UUID]int32{}
	for _, item := range items {
		productAmounts[item.ProductID] += item.Quantity
	}

	for id, amount := range productAmounts {
		err = qtx.ReduceProductInventory(ctx, repo.ReduceProductInventoryParams{ID: id, Amount: amount})
		if err != nil {
			return err
		}
	}

	return err
}

// Decline updates order status and refunds
func Decline(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	order, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusDeclined)
	if err != nil {
		return err
	}

	customerWallet, err := qtx.GetWalletForUser(ctx, order.CustomerID)
	if err != nil {
		return err
	}

	if _, err := qtx.AddWalletBalance(
		ctx,
		repo.AddWalletBalanceParams{
			ID:     customerWallet.ID,
			Amount: currency.AddFee(order.TotalPricePico), // Return the whole paid amount
		}); err != nil {
		return err
	}

	return nil
}

type DispatchParams struct {
	OrderID uuid.UUID
	Info    string
}

func Dispatch(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusDispatched)
	return err
}

func Finalize(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	order, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusFinalized)
	if err != nil {
		return err
	}

	vendor, err := qtx.GetVendorForOrder(ctx, order.ID)
	if err != nil {
		return err
	}

	wallet, err := qtx.GetWalletForUser(ctx, vendor.ID)
	if err != nil {
		return err
	}

	_, err = qtx.AddWalletBalance(
		ctx,
		repo.AddWalletBalanceParams{
			ID:     wallet.ID,
			Amount: order.TotalPricePico,
		})
	if err != nil {
		return err
	}

	return nil
}

// Is AF-time extending available
func ExtendAvailable(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) (bool, error) {
	o, err := qtx.GetOrder(ctx, orderID)
	if err != nil {
		return false, err
	}

	sinceDispatch := time.Since(o.DispatchedAt)
	if o.NumExtends == 0 && sinceDispatch < config.ExtendUnavailableWindow {
		return false, ErrExtendUnavailable
	} else if o.NumExtends == 1 && sinceDispatch < config.ExtendUnavailableWindow+config.OrderDeliveryWindow {
		return false, ErrReExtendUnavailable
	}

	return true, nil
}

// Extend AF-time
func Extend(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := qtx.ExtendOrder(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUnableToExtendFurther
	}
	return err
}
