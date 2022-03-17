package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *roleResolver) Users(ctx context.Context, obj *dbmodels.Role) ([]*dbmodels.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *roleResolver) Teams(ctx context.Context, obj *dbmodels.Role) ([]*dbmodels.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *teamResolver) Users(ctx context.Context, obj *dbmodels.Team) (*model.Users, error) {
	users := make([]*dbmodels.User, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return &model.Users{
		Nodes: users,
	}, nil
}

func (r *userResolver) Teams(ctx context.Context, obj *dbmodels.User) (*model.Teams, error) {
	teams := make([]*dbmodels.Team, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Teams").Find(&teams)
	if err != nil {
		return nil, err
	}
	return &model.Teams{
		Nodes: teams,
	}, nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type roleResolver struct{ *Resolver }
type teamResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
