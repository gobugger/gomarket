package vendor

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
	"slices"
)

type UpdateInfoParams repo.UpdateVendorInfoParams

func UpdateInfo(ctx context.Context, qtx *repo.Queries, p UpdateInfoParams) error {
	l, err := qtx.GetVendorLicenseForUser(ctx, p.ID)
	if err != nil {
		return err
	}

	_, err = qtx.UpdateVendorInfo(ctx, repo.UpdateVendorInfoParams{ID: l.ID, VendorInfo: p.VendorInfo})
	return err
}

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
