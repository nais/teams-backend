package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type User struct {
	*sqlc.User
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

func (d *database) GetUserByApiKey(ctx context.Context, apiKey string) (*User, error) {
	user, err := d.querier.GetUserByApiKey(ctx, apiKey)
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

func (d *database) RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error {
	return d.querier.RemoveAllUserRoles(ctx, userID)
}

func (d *database) RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error {
	return d.querier.RemoveApiKeysFromUser(ctx, userID)
}

func (d *database) SetUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error) {
	user, err := d.querier.SetUserName(ctx, sqlc.SetUserNameParams{
		Name: name,
		ID:   userID,
	})
	if err != nil {
		return nil, err
	}

	return &User{User: user}, nil
}

func (d *database) GetUsersByEmail(ctx context.Context, email string) ([]*User, error) {
	users, err := d.querier.GetUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return d.getUsers(users)
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
