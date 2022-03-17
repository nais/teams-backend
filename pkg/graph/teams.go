package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/auth"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*dbmodels.Team, error) {
	err := auth.Allowed(ctx, r.console, auth.AccessReadWrite, "teams", "createTeam")
	if err != nil {
		return nil, err
	}

	// New team
	team := &dbmodels.Team{
		Slug:    &input.Slug,
		Name:    &input.Name,
		Purpose: input.Purpose,
	}

	// Assign creator as admin for team
	role := &dbmodels.Role{
		System:      r.console,
		Resource:    fmt.Sprintf("teams:%s", input.Slug),
		AccessLevel: auth.AccessReadWrite,
		Permission:  auth.PermissionAllow,
		User:        auth.UserFromContext(ctx),
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		tx.Create(team)
		tx.Create(role)
		return tx.Error
	})

	if err != nil {
		return nil, err
	}

	r.trigger <- team
	return team, nil
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

	// all models populated, check ACL now
	err := auth.Allowed(ctx, r.console, auth.AccessReadWrite, "teams", "teams:"+*team.Slug)
	if err != nil {
		return nil, err
	}

	err = r.db.Model(team).Association("Users").Append(users)
	if err != nil {
		return nil, err
	}
	tx.Preload("Users").First(team)
	r.trigger <- team
	return team, nil
}

func (r *queryResolver) Teams(ctx context.Context) (*model.Teams, error) {
	// all models populated, check ACL now
	err := auth.Allowed(ctx, r.console, auth.AccessRead, "teams")
	if err != nil {
		return nil, err
	}

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
