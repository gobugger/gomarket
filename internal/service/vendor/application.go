package vendor

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/service/settings"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/internal/util/db"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/minio/minio-go/v7"
	"io"
)

type CreateApplicationParams struct {
	UserID         uuid.UUID
	Logo           io.ReadSeeker
	InventoryImage io.ReadSeeker
	ExistingVendor bool
	Letter         string
}

func CreateApplication(ctx context.Context, qtx *repo.Queries, mc *minio.Client, p CreateApplicationParams) (repo.VendorApplication, error) {
	settings, err := settings.Get(ctx, qtx)
	if err != nil {
		return repo.VendorApplication{}, err
	}

	applicationPrice := settings.VendorApplicationPrice

	if p.ExistingVendor {
		applicationPrice = 0
	} else {
		wallet, err := qtx.GetWalletForUser(ctx, p.UserID)
		if err != nil {
			return repo.VendorApplication{}, err
		}

		if wallet.BalancePico < applicationPrice {
			return repo.VendorApplication{}, ErrNotEnoughBalance
		}

		_, err = qtx.ReduceWalletBalance(ctx, repo.ReduceWalletBalanceParams{ID: wallet.ID, Amount: applicationPrice})
		if err != nil {
			return repo.VendorApplication{}, err
		}
	}

	application, err := qtx.CreateVendorApplication(
		ctx,
		repo.CreateVendorApplicationParams{
			ExistingVendor: p.ExistingVendor,
			Letter:         p.Letter,
			PricePaidPico:  applicationPrice,
			UserID:         p.UserID})
	if err != nil {
		if db.ErrCode(err) == pgerrcode.UniqueViolation {
			return repo.VendorApplication{}, ErrUserHasAlreadyApplied
		}
		return repo.VendorApplication{}, err
	}

	logo, err := util.DecodeImage(p.Logo)
	if err != nil {
		return repo.VendorApplication{}, err
	}

	logo = util.TransformImage(logo, 50, 50)

	if err := util.SaveImage(ctx, mc, "logo", p.UserID.String(), logo); err != nil {
		return repo.VendorApplication{}, err
	}

	if !p.ExistingVendor {
		inventory, err := util.DecodeImage(p.InventoryImage)
		if err != nil {
			return repo.VendorApplication{}, err
		}

		if err := util.SaveImage(ctx, mc, "inventory", application.ID.String(), inventory); err != nil {
			return repo.VendorApplication{}, err
		}
	}

	return application, err
}

func AcceptApplication(ctx context.Context, qtx *repo.Queries, applicationID uuid.UUID) (repo.VendorLicense, error) {
	application, err := qtx.GetVendorApplication(ctx, applicationID)
	if err != nil {
		return repo.VendorLicense{}, err
	}

	license, err := qtx.CreateVendorLicense(ctx, repo.CreateVendorLicenseParams{
		PricePaidPico: application.PricePaidPico,
		UserID:        application.UserID,
	})
	if err != nil {
		return license, err
	}

	if err := qtx.DeleteVendorApplication(ctx, applicationID); err != nil {
		return license, err
	}

	_, err = qtx.CreateTermsOfService(ctx, repo.CreateTermsOfServiceParams{
		Content:  "",
		VendorID: application.UserID,
	})

	return license, err
}

func DeclineApplication(ctx context.Context, qtx *repo.Queries, applicationID uuid.UUID) error {
	return qtx.DeleteVendorApplication(ctx, applicationID)
}
