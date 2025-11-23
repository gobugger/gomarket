package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
)

type Vendor struct {
	User               repo.User
	License            repo.VendorLicense
	Rating             Rating
	NumReviews         int
	NumPendingDisputes int
	NumOrdersCompleted int
	NumOrdersInEscrow  int
}

type VendorView struct{}

func (pv VendorView) Get(ctx context.Context, q *repo.Queries, vendorID uuid.UUID) (Vendor, error) {
	user, err := q.GetUser(ctx, vendorID)
	if err != nil {
		return Vendor{}, err
	}

	license, err := q.GetVendorLicenseForUser(ctx, vendorID)
	if err != nil {
		return Vendor{}, err
	}

	reviews, err := V.Review.GetAllForVendor(ctx, q, vendorID)
	if err != nil {
		return Vendor{}, err
	}

	orders, err := q.GetOrdersForVendor(ctx, vendorID)
	if err != nil {
		return Vendor{}, err
	}

	numPendingDisputes := 0
	numOrdersInEscrow := 0
	numOrdersCompleted := 0
	for _, order := range orders {
		switch order.Status {
		case repo.OrderStatusDisputed:
			numPendingDisputes++
		case repo.OrderStatusPaid, repo.OrderStatusAccepted, repo.OrderStatusDispatched:
			numOrdersInEscrow++
		case repo.OrderStatusFinalized:
			numOrdersCompleted++
		}
	}

	return Vendor{
		User:               user,
		License:            license,
		Rating:             calculateRating2(reviews),
		NumReviews:         len(reviews),
		NumPendingDisputes: numPendingDisputes,
		NumOrdersCompleted: numOrdersCompleted,
		NumOrdersInEscrow:  numOrdersInEscrow,
	}, nil
}
