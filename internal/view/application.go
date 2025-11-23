package view

import (
	"context"
	"github.com/google/uuid"
	"github.com/gobugger/gomarket/internal/repo"
)

type VendorApplication struct {
	Application repo.VendorApplication
	User        repo.User
}

type VendorApplicationView struct{}

func (v VendorApplicationView) Get(ctx context.Context, q *repo.Queries, id uuid.UUID) (*VendorApplication, error) {
	application, err := q.GetVendorApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	user, err := q.GetUser(ctx, application.UserID)
	if err != nil {
		return nil, err
	}

	return &VendorApplication{
		Application: application,
		User:        user,
	}, nil
}
