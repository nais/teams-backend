package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

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

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
