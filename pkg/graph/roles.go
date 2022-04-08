package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) AssignRoleToUser(ctx context.Context, input model.AssignRoleInput) (*dbmodels.RoleBinding, error) {
	rb := &dbmodels.RoleBinding{
		UserID: input.UserID,
		RoleID: input.RoleID,
		TeamID: input.TeamID,
	}

	tx := r.db.WithContext(ctx).Create(rb)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return rb, nil
}

func (r *mutationResolver) RemoveRoleFromUser(ctx context.Context, input model.AssignRoleInput) (bool, error) {
	rb := &dbmodels.RoleBinding{}

	tx := r.db.WithContext(ctx).First(rb, "user_id = ? AND team_id = ? AND role_id = ?", input.UserID, input.TeamID, input.RoleID).Delete(rb)
	if tx.Error != nil {
		return false, tx.Error
	}
	return true, nil
}

func (r *queryResolver) Roles(ctx context.Context) ([]*dbmodels.Role, error) {
	roles := make([]*dbmodels.Role, 0)
	tx := r.db.WithContext(ctx).Find(&roles)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roles, nil
}
