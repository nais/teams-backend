package db

import (
	"context"

	"github.com/google/uuid"
)

func (d *database) CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	serviceAccount, err := d.querier.CreateServiceAccount(ctx, name)
	if err != nil {
		return nil, err
	}

	return &ServiceAccount{ServiceAccount: serviceAccount}, nil
}

func (d *database) GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error) {
	serviceAccount, err := d.querier.GetServiceAccountByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return &ServiceAccount{ServiceAccount: serviceAccount}, nil
}

func (d *database) GetServiceAccountByApiKey(ctx context.Context, apiKey string) (*ServiceAccount, error) {
	serviceAccount, err := d.querier.GetServiceAccountByApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return &ServiceAccount{ServiceAccount: serviceAccount}, nil
}

func (d *database) GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error) {
	rows, err := d.querier.GetServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	serviceAccounts := make([]*ServiceAccount, 0)
	for _, row := range rows {
		serviceAccounts = append(serviceAccounts, &ServiceAccount{ServiceAccount: row})
	}

	return serviceAccounts, nil
}

func (d *database) DeleteServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error {
	return d.querier.DeleteServiceAccount(ctx, serviceAccountID)
}
