package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
)

func (r *teamResolver) Users(ctx context.Context, obj *dbmodels.Team) ([]*dbmodels.User, error) {
	users := make([]*dbmodels.User, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userResolver) Teams(ctx context.Context, obj *dbmodels.User) ([]*dbmodels.Team, error) {
	teams := make([]*dbmodels.Team, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Teams").Find(&teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *userResolver) Roles(ctx context.Context, obj *dbmodels.User, teamID *uuid.UUID) ([]*dbmodels.Role, error) {
	roleBindings := make([]*dbmodels.RoleBinding, 0)
	roles := make([]*dbmodels.Role, 0)

	tx := r.db.WithContext(ctx).Model(&dbmodels.RoleBinding{}).Preload("Role").Find(&roleBindings, "user_id = ? AND team_id = ?", obj.ID, *teamID)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, rb := range roleBindings {
		roles = append(roles, rb.Role)
	}

	return roles, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type teamResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
