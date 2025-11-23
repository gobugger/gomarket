package market

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/auth"
	"github.com/gobugger/gomarket/internal/service/currency"
	order_service "github.com/gobugger/gomarket/internal/service/order"
	"github.com/gobugger/gomarket/internal/service/payment"
	"github.com/gobugger/gomarket/internal/service/payment/wallet"
	"github.com/gobugger/gomarket/internal/worker"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type CreateOrderParams struct {
	CustomerID       uuid.UUID
	DeliveryMethodID uuid.UUID
	Details          string
	UseWallet        bool
}

// CreateOrder creates an order and bills from wallet or creates an invoice for the order
func CreateOrder(ctx context.Context, tx pgx.Tx, rc *river.Client[pgx.Tx], p CreateOrderParams) (repo.Order, error) {
	qtx := repo.New(tx)

	order, err := order_service.Create(ctx, qtx, order_service.CreateParams{
		CustomerID:       p.CustomerID,
		DeliveryMethodID: p.DeliveryMethodID,
		Details:          p.Details,
	})
	if err != nil {
		return order, err
	}

	totalBill := currency.AddFee(order.TotalPricePico)

	if p.UseWallet {
		wallet, err := qtx.GetWalletForUser(ctx, p.CustomerID)
		if err != nil {
			return order, err
		}

		if wallet.BalancePico < totalBill {
			return order, order_service.ErrNotEnoughBalance
		}

		_, err = qtx.ReduceWalletBalance(ctx, repo.ReduceWalletBalanceParams{
			ID:     wallet.ID,
			Amount: totalBill,
		})
		if err != nil {
			return order, err
		}

		order, err = order_service.UpdateStatus(ctx, qtx, order.ID, repo.OrderStatusPaid)
		if err != nil {
			return order, err
		}
	} else {
		invoice, err := payment.CreateInvoice(ctx, qtx, totalBill)
		if err != nil {
			return order, err
		}

		_, err = rc.InsertTx(ctx, tx, worker.PrepareInvoiceArgs{ID: invoice.ID}, nil)
		if err != nil {
			return order, err
		}

		_, err = qtx.CreateOrderInvoice(ctx, repo.CreateOrderInvoiceParams{
			OrderID:   order.ID,
			InvoiceID: invoice.ID,
		})
		if err != nil {
			return order, err
		}
	}

	return order, nil
}

type RegisterUserParams struct {
	Username string
	Password string
}

func RegisterUser(ctx context.Context, tx pgx.Tx, rc *river.Client[pgx.Tx], p RegisterUserParams) (repo.User, error) {
	qtx := repo.New(tx)

	user, err := auth.Register(ctx, qtx, auth.RegisterParams{Username: p.Username, Password: p.Password})
	if err != nil {
		return user, err
	}

	w, err := wallet.CreateWallet(ctx, qtx, user.ID)
	if err != nil {
		return user, err
	}

	invoice, err := qtx.GetInvoiceForWallet(ctx, w.ID)
	if err != nil {
		return user, err
	}

	// Invoice is created for users deposit so invoke PrepareInvoices
	_, err = rc.InsertTx(ctx, tx, worker.PrepareInvoiceArgs{ID: invoice.ID}, nil)
	return user, err
}
