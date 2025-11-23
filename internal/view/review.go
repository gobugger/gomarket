package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
	"time"
)

type Review struct {
	Review repo.Review
	Author repo.User
}

type ProductReview struct {
	Review          repo.ProductReview
	TotalQuantity   int32
	TotalPriceCents int64
	AuthorName      string
	CreatedAt       time.Time
}

type ReviewView struct{}

func (rv ReviewView) GetAllForProduct(ctx context.Context, q *repo.Queries, productID uuid.UUID) ([]ProductReview, error) {
	rows, err := q.GetReviewsForProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	reviews := make([]ProductReview, 0, len(rows))
	for _, row := range rows {
		reviews = append(reviews, ProductReview{
			Review:          row.ProductReview,
			TotalQuantity:   row.TotalQuantity,
			TotalPriceCents: int64(row.TotalPriceCent),
			AuthorName:      row.AuthorName,
			CreatedAt:       row.CreatedAt,
		})
	}

	return reviews, nil
}

func (rv ReviewView) GetAllForVendor(ctx context.Context, q *repo.Queries, vendorID uuid.UUID) ([]Review, error) {
	rows, err := q.GetReviewsForVendor(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0, len(rows))
	for _, row := range rows {
		reviews = append(reviews, Review{
			Review: row.Review,
			Author: row.User,
		})
	}

	return reviews, nil
}
