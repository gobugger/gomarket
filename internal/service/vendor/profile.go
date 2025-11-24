package vendor

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/google/uuid"
	"slices"
)

type DeliveryMethod struct {
	Description string
	PriceCent   int64
}

type UpdateDeliveryMethodsParams struct {
	VendorID        uuid.UUID
	DeliveryMethods []DeliveryMethod
}

// Updates vendors delivery methods
func UpdateDeliveryMethods(ctx context.Context, qtx *repo.Queries, p UpdateDeliveryMethodsParams) error {
	dms, err := qtx.GetDeliveryMethodsForVendor(ctx, p.VendorID)
	if err != nil {
		return err
	}

	// Delete invalidated delivery methods
	for _, dm := range dms {
		if !slices.ContainsFunc(p.DeliveryMethods, func(newDm DeliveryMethod) bool {
			return dm.Description == newDm.Description && dm.PriceCent == newDm.PriceCent
		}) {
			if err := qtx.DeleteDeliveryMethod(ctx, dm.ID); err != nil {
				return err
			}
		}
	}

	// Create new/updated delivery methods
	for _, dm := range p.DeliveryMethods {
		if !slices.ContainsFunc(dms, func(newDm repo.DeliveryMethod) bool {
			return dm.Description == newDm.Description && dm.PriceCent == newDm.PriceCent
		}) {
			_, err := qtx.CreateDeliveryMethod(ctx, repo.CreateDeliveryMethodParams{
				Description: dm.Description,
				PriceCent:   dm.PriceCent,
				VendorID:    p.VendorID,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
