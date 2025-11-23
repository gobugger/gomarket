package product

import (
	"context"
	"fmt"
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"io"
)

type PriceTier struct {
	Quantity  int32
	PriceCent int64
}

type CreateParams struct {
	Title       string
	Description string
	CategoryID  uuid.UUID
	Inventory   int32
	ShipsFrom   string
	ShipsTo     string
	VendorID    uuid.UUID
	Image       io.ReadSeeker
	PriceTiers  []PriceTier
}

func Create(ctx context.Context, qtx *repo.Queries, mc *minio.Client, p CreateParams) (repo.Product, error) {
	product, err := qtx.CreateProduct(
		ctx,
		repo.CreateProductParams{
			Title:       p.Title,
			Description: p.Description,
			CategoryID:  p.CategoryID,
			Inventory:   p.Inventory,
			ShipsFrom:   p.ShipsFrom,
			ShipsTo:     p.ShipsTo,
			VendorID:    p.VendorID,
		})
	if err != nil {
		return repo.Product{}, err
	}

	for _, pt := range p.PriceTiers {
		_, err := qtx.CreatePriceTier(
			ctx,
			repo.CreatePriceTierParams{
				ProductID: product.ID,
				Quantity:  pt.Quantity,
				PriceCent: pt.PriceCent,
			})
		if err != nil {
			return repo.Product{}, err
		}
	}

	original, err := util.DecodeImage(p.Image)
	if err != nil {
		return repo.Product{}, fmt.Errorf("failed to decode product image: %w", err)
	}

	thumbnail := util.TransformImage(original, 200, 200)

	id := product.ID.String()

	if err = util.SaveImage(ctx, mc, "product", id, original); err != nil {
		return repo.Product{}, err
	}

	if err = util.SaveImage(ctx, mc, "thumbnail", id, thumbnail); err != nil {
		return repo.Product{}, err
	}

	return product, nil
}

func Delete(ctx context.Context, qtx *repo.Queries, id uuid.UUID) error {
	return qtx.DeleteProduct(ctx, id)
}

func UpdateInventory(ctx context.Context, qtx *repo.Queries, id uuid.UUID, inventory int32) error {
	if inventory < 0 {
		return fmt.Errorf("inventory can't be negative")
	}
	return qtx.UpdateProductInventory(ctx, repo.UpdateProductInventoryParams{ID: id, Inventory: inventory})
}
