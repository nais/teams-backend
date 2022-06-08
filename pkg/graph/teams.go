package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/roles"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*dbmodels.Team, error) {
	err := authz.Authorized(ctx, r.console, nil, authz.AccessLevelCreate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	user := authz.UserFromContext(ctx)

	team := &dbmodels.Team{
		Slug:    input.Slug,
		Name:    &input.Name,
		Purpose: input.Purpose,
		Users:   []*dbmodels.User{user},
	}

	tx := r.db.WithContext(ctx)

	err = tx.Transaction(func(tx *gorm.DB) error {
		teamEditor := &dbmodels.Role{
			Name: roles.TeamEditor,
		}

		tx.Where(teamEditor).First(teamEditor)
		if tx.Error != nil {
			return tx.Error
		}

		tx.Create(team)
		if tx.Error != nil {
			return tx.Error
		}

		roleBinding := &dbmodels.RoleBinding{
			TeamID: team.ID,
			Role:   teamEditor,
			User:   user,
		}

		return tx.Create(roleBinding).Error
	})

	if err != nil {
		return nil, err
	}

	// FIXME: Add log entry for team creation?

	team, err = r.teamWithAssociations(ctx, *team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch complete team: %s", err)
	}
	r.trigger <- team

	return team, nil
}

func (r *mutationResolver) AddUsersToTeam(ctx context.Context, input model.AddUsersToTeamInput) (*dbmodels.Team, error) {
	tx := r.db.WithContext(ctx)

	team := &dbmodels.Team{}
	err := tx.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = tx.Where(input.UserIds).Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	err = authz.Authorized(ctx, r.console, team, authz.AccessLevelUpdate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	err = tx.Model(team).Association("Users").Append(users)
	if err != nil {
		return nil, err
	}

	team, err = r.teamWithAssociations(ctx, *team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch complete team: %s", err)
	}
	r.trigger <- team

	return team, nil
}

func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, input model.RemoveUsersFromTeamInput) (*dbmodels.Team, error) {
	tx := r.db.WithContext(ctx)

	team := &dbmodels.Team{}
	err := tx.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = tx.Where(input.UserIds).Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	err = authz.Authorized(ctx, r.console, team, authz.AccessLevelUpdate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	err = tx.Model(team).Association("Users").Delete(users)
	if err != nil {
		return nil, err
	}

	// FIXME: Also remove all role bindings for the removed users that are attached to the specific team

	team, err = r.teamWithAssociations(ctx, *team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch complete team: %s", err)
	}
	r.trigger <- team

	return team, nil
}

func (r *queryResolver) Teams(ctx context.Context, pagination *model.Pagination, query *model.TeamsQuery, sort *model.TeamsSort) (*model.Teams, error) {
	err := authz.Authorized(ctx, r.console, nil, authz.AccessLevelRead, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	teams := make([]*dbmodels.Team, 0)
	if sort == nil {
		sort = &model.TeamsSort{
			Field:     model.TeamSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pageInfo, err := r.paginatedQuery(ctx, pagination, query, sort, &dbmodels.Team{}, &teams)
	return &model.Teams{
		PageInfo: pageInfo,
		Nodes:    teams,
	}, err
}

func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*dbmodels.Team, error) {
	tx := r.db.WithContext(ctx)
	team := &dbmodels.Team{}
	err := tx.Where("id = ?", id).First(team).Error
	if err != nil {
		return nil, err
	}

	err = authz.Authorized(ctx, r.console, team, authz.AccessLevelRead, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *teamResolver) Users(ctx context.Context, obj *dbmodels.Team) ([]*dbmodels.User, error) {
	users := make([]*dbmodels.User, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *teamResolver) Metadata(ctx context.Context, obj *dbmodels.Team) (map[string]interface{}, error) {
	metadata := make([]*dbmodels.TeamMetadata, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Metadata").Find(&metadata)
	if err != nil {
		return nil, err
	}

	kv := make(map[string]interface{})

	for _, pair := range metadata {
		kv[pair.Key] = pair.Value
	}

	return kv, nil
}

func (r *teamResolver) AuditLogs(ctx context.Context, obj *dbmodels.Team) ([]*dbmodels.AuditLog, error) {
	auditLogs := make([]*dbmodels.AuditLog, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("AuditLogs").Find(&auditLogs)
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
