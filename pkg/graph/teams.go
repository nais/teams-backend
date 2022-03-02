package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*dbmodels.Team, error) {
	u := &dbmodels.Team{
		Slug:    &input.Slug,
		Name:    &input.Name,
		Purpose: input.Purpose,
	}
	tx := r.db.WithContext(ctx).Create(u)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return u, nil
}

func (r *mutationResolver) AddUsersToTeam(ctx context.Context, input model.AddUsersToTeamInput) (*dbmodels.Team, error) {
	users := make([]*dbmodels.User, 0)
	team := &dbmodels.Team{}
	tx := r.db.WithContext(ctx)

	tx.Find(&users, input.UserID)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if len(users) != len(input.UserID) {
		return nil, fmt.Errorf("one or more non-existing user IDs given as parameters")
	}

	tx.First(team, "id = ?", input.TeamID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	err := r.db.Model(team).Association("Users").Append(users)
	if err != nil {
		return nil, err
	}
	tx.Preload("Users").First(team)
	return team, nil
}

func (r *queryResolver) Teams(ctx context.Context) (*model.Teams, error) {
	var count int64
	teams := make([]*dbmodels.Team, 0)
	tx := r.db.WithContext(ctx).Find(&teams)
	tx.Count(&count)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &model.Teams{
		Pagination: &model.Pagination{
			Results: int(count),
		},
		Nodes: teams,
	}, nil
}
