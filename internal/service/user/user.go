package user

import (
	"context"
	"github.com/gobugger/gomarket/internal/repo"
)

type UpdateSettingsParams repo.UpdateUserSettingsParams

func UpdateSettings(ctx context.Context, qtx *repo.Queries, p UpdateSettingsParams) error {
	return qtx.UpdateUserSettings(ctx, repo.UpdateUserSettingsParams{
		ID:               p.ID,
		Locale:           p.Locale,
		Currency:         p.Currency,
		TwofaEnabled:     p.TwofaEnabled,
		IncognitoEnabled: p.IncognitoEnabled,
	})
}
