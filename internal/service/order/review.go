package order

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
)

type ProductReview struct {
	Grade   int32
	Comment string
}

type CreateReviewParams struct {
	OrderID        uuid.UUID
	Grade          int32
	Comment        string
	ProductReviews map[uuid.UUID]ProductReview
}

func CreateReview(ctx context.Context, qtx *repo.Queries, p CreateReviewParams) error {
	_, err := qtx.CreateReview(
		ctx,
		repo.CreateReviewParams{
			Grade:   p.Grade,
			Comment: p.Comment,
			OrderID: p.OrderID,
		})
	if err != nil {
		return err
	}

	items, err := qtx.GetOrderItems(ctx, p.OrderID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if pr, ok := p.ProductReviews[item.ID]; ok {
			_, err := qtx.CreateProductReview(ctx, repo.CreateProductReviewParams{
				OrderItemID: item.ID,
				Grade:       pr.Grade,
				Comment:     pr.Comment,
			})
			if err != nil {
				return err
			}
		}
	}

	return err
}

func CreateDefaultReview(ctx context.Context, qtx *repo.Queries, orderID uuid.UUID) error {
	_, err := qtx.CreateReview(
		ctx,
		repo.CreateReviewParams{
			Grade:   5,
			Comment: "",
			OrderID: orderID,
		})
	if err != nil {
		return err
	}

	items, err := qtx.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err := qtx.CreateProductReview(ctx, repo.CreateProductReviewParams{
			OrderItemID: item.ID,
			Grade:       5,
			Comment:     "",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
