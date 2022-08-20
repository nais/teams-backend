package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type User struct {
	*sqlc.User
	Roles []*Role
	Teams []*Team
}

type Role struct {
	*sqlc.UserRole
	Name           sqlc.RoleName
	Authorizations []sqlc.AuthzName
}

func (d *database) AddUser(ctx context.Context, name, email string) (*User, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	user, err := d.querier.CreateUser(ctx, sqlc.CreateUserParams{
		ID:    id,
		Name:  name,
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := d.querier.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return d.getUser(ctx, &User{User: user})
}

func (d *database) GetUserByApiKey(ctx context.Context, apiKey string) (*User, error) {
	user, err := d.querier.GetUserByApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return d.getUser(ctx, &User{User: user})
}

func (d *database) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := d.querier.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return d.getUser(ctx, &User{User: user})
}

func (d *database) getUser(ctx context.Context, user *User) (*User, error) {
	userRoles, err := d.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	user.Roles = userRoles

	return user, nil
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
			Name:           userRole.RoleName,
		})
	}

	return userRoles, nil
}
