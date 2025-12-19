package settings

import (
	"context"
	"encoding/json"
	"github.com/gobugger/gomarket/internal/repo"
	"math/big"
)

type Settings struct {
	VendorApplicationPrice *big.Int `json:"vendor_application_price"`
}

func Get(ctx context.Context, q *repo.Queries) (Settings, error) {
	data, err := q.GetSettings(ctx)
	if err != nil {
		return Settings{}, err
	}

	settings := Settings{}
	err = json.Unmarshal(data, &settings)
	return settings, err
}

func Set(ctx context.Context, q *repo.Queries, settings Settings) error {
	data, err := json.Marshal(&settings)
	if err != nil {
		return err
	}

	return q.UpdateSettings(ctx, data)
}
