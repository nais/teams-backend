package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/models"
)

func (r *teamResolver) Users(ctx context.Context, obj *models.Team) (*model.Users, error) {
	users := make([]*models.User, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return &model.Users{
		Meta: &model.Meta{
			NumResults: len(users),
		},
		Nodes: users,
	}, nil
}

func (r *userResolver) Teams(ctx context.Context, obj *models.User) (*model.Teams, error) {
	teams := make([]*models.Team, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Teams").Find(&teams)
	if err != nil {
		return nil, err
	}
	return &model.Teams{
		Meta: &model.Meta{
			NumResults: len(teams),
		},
		Nodes: teams,
	}, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type teamResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
