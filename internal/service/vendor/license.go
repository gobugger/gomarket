package vendor

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
)

var (
	ErrNotEnoughBalance      = errors.New("not enough balance")
	ErrUserIsAlreadyVendor   = errors.New("user is already a vendor")
	ErrUserHasAlreadyApplied = errors.New("user has already applied")
)

func HasLicense(ctx context.Context, qtx *repo.Queries, userID uuid.UUID) bool {
	val, err := qtx.HasVendorLicense(ctx, userID)
	return err == nil && val == 1
}
