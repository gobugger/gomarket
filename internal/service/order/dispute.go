package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"math/big"
)

func Dispute(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusDisputed)
	return err
}

type CreateDisputeOfferParams struct {
	OrderID      uuid.UUID
	RefundFactor float64
}

func CreateDisputeOffer(ctx context.Context, qtx *repo.Queries, p CreateDisputeOfferParams) (repo.DisputeOffer, error) {
	if p.RefundFactor < 0 || p.RefundFactor > 1 {
		return repo.DisputeOffer{}, fmt.Errorf("refund factor must be in [0, 1] for offers")
	}

	offer, err := qtx.CreateDisputeOffer(ctx, repo.CreateDisputeOfferParams{OrderID: p.OrderID, RefundFactor: p.RefundFactor})
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.DisputeOffer{}, ErrInvalidStatus
	}

	return offer, err
}

func AcceptDisputeOffer(ctx context.Context, qtx *repo.Queries, offerID uuid.UUID) error {
	offer, err := qtx.UpdateDisputeOfferStatus(ctx, repo.UpdateDisputeOfferStatusParams{ID: offerID, Status: repo.DisputeOfferStatusAccepted})
	if err != nil {
		return err
	}

	return resolveDispute(ctx, qtx, offer.OrderID, offer.RefundFactor)
}

func DeclineDisputeOffer(ctx context.Context, qtx *repo.Queries, offerID uuid.UUID) error {
	_, err := qtx.UpdateDisputeOfferStatus(ctx, repo.UpdateDisputeOfferStatusParams{ID: offerID, Status: repo.DisputeOfferStatusDeclined})
	return err
}

type ForceResolveDisputeParams struct {
	OrderID      uuid.UUID
	RefundFactor float64
}

func ForceResolveDispute(ctx context.Context, qtx *repo.Queries, p ForceResolveDisputeParams) error {
	offer, err := qtx.CreateDisputeOfferWithStatus(ctx, repo.CreateDisputeOfferWithStatusParams{
		RefundFactor: p.RefundFactor,
		OrderID:      p.OrderID,
		Status:       repo.DisputeOfferStatusForced,
	})
	if err != nil {
		return err
	}

	return resolveDispute(ctx, qtx, offer.OrderID, offer.RefundFactor)
}

func resolveDispute(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID, refundFactor float64) error {
	if refundFactor < 0 || refundFactor > 1 {
		return fmt.Errorf("refund factor must be in [0, 1]")
	}

	order, err := UpdateStatus(ctx, qtx, orderID, repo.OrderStatusSettled)
	if err != nil {
		return err
	}

	vendor, err := qtx.GetVendorForOrder(ctx, orderID)
	if err != nil {
		return err
	}

	totalRefund := repo.Num2Big(order.TotalPricePico)
	customerRefund, _ := new(big.Float).Mul(new(big.Float).SetInt(totalRefund), big.NewFloat(refundFactor)).Int(nil)
	vendorRefund := new(big.Int).Sub(totalRefund, customerRefund)

	if customerRefund.Sign() > 0 {
		customerWallet, err := qtx.GetWalletForUser(ctx, order.CustomerID)
		if err != nil {
			return err
		}

		if _, err := qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: customerWallet.ID, Amount: repo.Big2Num(customerRefund)}); err != nil {
			return err
		}
	}
	if vendorRefund.Sign() > 0 {
		vendorWallet, err := qtx.GetWalletForUser(ctx, vendor.ID)
		if err != nil {
			return err
		}

		if _, err := qtx.AddWalletBalance(ctx, repo.AddWalletBalanceParams{ID: vendorWallet.ID, Amount: repo.Big2Num(vendorRefund)}); err != nil {
			return err
		}
	}

	return err
}
