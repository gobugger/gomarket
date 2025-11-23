package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
)

type Product struct {
	repo.Product
	PriceTiers      []repo.PriceTier
	DeliveryMethods []repo.DeliveryMethod
	Vendor          Vendor
	Reviews         []ProductReview
	Rating          Rating
	NumReviews      int
}

type ProductPage struct {
	Products []Product
	Page     int
	NumPages int
}

type ProductView struct{}

func (pv ProductView) Get(ctx context.Context, q *repo.Queries, productID uuid.UUID) (Product, error) {
	product, err := q.GetProduct(ctx, productID)
	if err != nil {
		return Product{}, err
	}

	prices, err := q.GetPriceTiers(ctx, productID)
	if err != nil {
		return Product{}, err
	}

	dms, err := q.GetDeliveryMethodsForProduct(ctx, productID)
	if err != nil {
		return Product{}, err
	}

	vendor, err := V.Vendor.Get(ctx, q, product.VendorID)
	if err != nil {
		return Product{}, err
	}

	reviews, err := V.Review.GetAllForProduct(ctx, q, productID)
	if err != nil {
		return Product{}, err
	}

	return Product{
		Product:         product,
		PriceTiers:      prices,
		DeliveryMethods: dms,
		Vendor:          vendor,
		Reviews:         reviews,
		Rating:          calculateRating(reviews),
		NumReviews:      len(reviews),
	}, nil
}

func (pv ProductView) GetAll(ctx context.Context, q *repo.Queries) ([]Product, error) {
	products, err := q.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	prices, err := q.GetAllPriceTiers(ctx)
	if err != nil {
		return nil, err
	}

	dms, err := q.GetAllDeliveryMethods(ctx)
	if err != nil {
		return nil, err
	}

	pricesForProduct := map[uuid.UUID][]repo.PriceTier{}
	dmsForProduct := map[uuid.UUID][]repo.DeliveryMethod{}
	reviewsForProduct := map[uuid.UUID][]ProductReview{}

	for _, price := range prices {
		pricesForProduct[price.ProductID] = append(pricesForProduct[price.ProductID], price)
	}

	for _, p := range products {
		for _, dm := range dms {
			if dm.VendorID == p.VendorID {
				dmsForProduct[p.ID] = append(dmsForProduct[p.ID], dm)
			}
		}

		reviews, err := V.Review.GetAllForProduct(ctx, q, p.ID)
		if err != nil {
			return nil, err
		}

		reviewsForProduct[p.ID] = append(reviewsForProduct[p.ID], reviews...)
	}

	res := make([]Product, 0, len(products))
	for _, product := range products {
		vendor, err := V.Vendor.Get(ctx, q, product.VendorID)
		if err != nil {
			return nil, err
		}

		reviews := reviewsForProduct[product.ID]

		res = append(res, Product{
			Product:         product,
			PriceTiers:      pricesForProduct[product.ID],
			DeliveryMethods: dmsForProduct[product.ID],
			Vendor:          vendor,
			Reviews:         reviews,
			Rating:          calculateRating(reviews),
			NumReviews:      len(reviews),
		})
	}
	return res, nil
}
