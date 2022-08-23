package db

import (
	"context"

	"github.com/nais/console/pkg/slug"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type User struct {
	*sqlc.User
	Roles []*Role
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

func (d *database) AddServiceAccount(ctx context.Context, name slug.Slug, email string, userID uuid.UUID) (*User, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	tx, err := d.connPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	querier := d.querier.WithTx(tx)
	serviceAccount, err := querier.CreateUser(ctx, sqlc.CreateUserParams{
		ID:    id,
		Name:  string(name),
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	err = querier.AddTargetedUserRole(ctx, sqlc.AddTargetedUserRoleParams{
		UserID:   userID,
		RoleName: sqlc.RoleNameServiceaccountowner,
		TargetID: nullUUID(&serviceAccount.ID),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &User{User: serviceAccount}, nil
}

func (d *database) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return d.querier.DeleteUser(ctx, userID)
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

	return d.getUser(ctx, &User{User: user})
}

func (d *database) GetUsersByEmail(ctx context.Context, email string) ([]*User, error) {
	users, err := d.querier.GetUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return d.getUsers(ctx, users)
}

func (d *database) GetUsers(ctx context.Context) ([]*User, error) {
	users, err := d.querier.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return d.getUsers(ctx, users)
}

func (d *database) getUsers(ctx context.Context, users []*sqlc.User) ([]*User, error) {
	result := make([]*User, 0)
	for _, user := range users {
		u, err := d.getUser(ctx, &User{User: user})
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

func (d *database) getUser(ctx context.Context, user *User) (*User, error) {
	userRoles, err := d.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	user.Roles = userRoles

	return user, nil
}

func (d *database) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	return d.getUserRoles(ctx, userID)
}
