package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) CreateUser(ctx context.Context, name, email, externalID string) (*User, error) {
	user, err := d.querier.CreateUser(ctx, sqlc.CreateUserParams{
		Name:       name,
		Email:      email,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return d.querier.DeleteUser(ctx, userID)
}

func (d *database) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := d.querier.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := d.querier.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetUserByExternalID(ctx context.Context, externalID string) (*User, error) {
	user, err := d.querier.GetUserByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) getUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	ur, err := d.querier.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	userRoles := make([]*Role, 0)
	for _, userRole := range ur {
		authorizations, err := d.querier.GetRoleAuthorizations(ctx, userRole.RoleName)
		if err != nil {
			return nil, err
		}

		userRoles = append(userRoles, &Role{
			UserRole:       userRole,
			Authorizations: authorizations,
		})
	}

	return userRoles, nil
}

func (d *database) RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error {
	return d.querier.RemoveAllUserRoles(ctx, userID)
}

func (d *database) UpdateUser(ctx context.Context, userID uuid.UUID, name, email, externalID string) (*User, error) {
	user, err := d.querier.UpdateUser(ctx, sqlc.UpdateUserParams{
		Email:      email,
		ExternalID: externalID,
		ID:         userID,
		Name:       name,
	})
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetUsers(ctx context.Context) ([]*User, error) {
	users, err := d.querier.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return d.getUsers(users)
}

func (d *database) getUsers(users []*sqlc.User) ([]*User, error) {
	result := make([]*User, 0)
	for _, user := range users {
		result = append(result, &User{User: user})
	}

	return result, nil
}

func (d *database) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	return d.getUserRoles(ctx, userID)
}
