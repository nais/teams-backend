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
	"github.com/nais/console/pkg/reconcilers"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/roles"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*dbmodels.Team, error) {
	err := authz.Authorized(ctx, r.system, nil, authz.AccessLevelCreate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	user := authz.UserFromContext(ctx)
	corr := &dbmodels.Correlation{}
	team := &dbmodels.Team{
		Slug:    *input.Slug,
		Name:    input.Name,
		Purpose: input.Purpose,
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		teamEditor := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.TeamEditor).First(teamEditor).Error
		if err != nil {
			return err
		}

		err = r.createTrackedObject(ctx, team)
		if err != nil {
			return err
		}

		ut := &dbmodels.UsersTeams{
			UserID: *user.ID,
			TeamID: *team.ID,
		}

		err = r.createTrackedObject(ctx, ut)
		if err != nil {
			return err
		}

		roleBinding := &dbmodels.RoleBinding{
			RoleID: *teamEditor.ID,
			TeamID: team.ID,
			UserID: *user.ID,
		}

		return r.createTrackedObject(ctx, roleBinding)
	})

	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpCreateTeam, *corr, *r.system, user, team, nil, "Team created")

	team, err = r.teamWithAssociations(*team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch team: %w", err)
	}

	r.teamReconciler <- reconcilers.Input{
		Corr: *corr,
		Team: *team,
	}

	return team, nil
}

func (r *mutationResolver) AddUsersToTeam(ctx context.Context, input model.AddUsersToTeamInput) (*dbmodels.Team, error) {
	team := &dbmodels.Team{}
	err := r.db.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = r.db.Where(input.UserIds).Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	err = authz.Authorized(ctx, r.system, team, authz.AccessLevelUpdate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		for _, userId := range input.UserIds {
			tm := &dbmodels.UsersTeams{
				UserID: *userId,
				TeamID: *team.ID,
			}
			err = r.createTrackedObjectIgnoringDuplicates(ctx, tm)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	team, err = r.teamWithAssociations(*team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch team: %w", err)
	}

	r.teamReconciler <- reconcilers.Input{
		Corr: *corr,
		Team: *team,
	}

	return team, nil
}

func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, input model.RemoveUsersFromTeamInput) (*dbmodels.Team, error) {
	team := &dbmodels.Team{}
	err := r.db.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = r.db.Where(input.UserIds).Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	err = authz.Authorized(ctx, r.system, team, authz.AccessLevelUpdate, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		err = tx.Where("user_id IN (?) AND team_id = ?", input.UserIds, team.ID).Delete(&dbmodels.UsersTeams{}).Error
		if err != nil {
			return err
		}

		err = tx.Where("user_id IN (?) AND team_id = ?", input.UserIds, team.ID).Delete(&dbmodels.RoleBinding{}).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	team, err = r.teamWithAssociations(*team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch team: %w", err)
	}
	r.teamReconciler <- reconcilers.Input{
		Corr: *corr,
		Team: *team,
	}

	return team, nil
}

func (r *mutationResolver) SynchronizeTeam(ctx context.Context, teamID *uuid.UUID) (bool, error) {
	team := &dbmodels.Team{}
	err := r.db.Where("id = ?", teamID).First(team).Error
	if err != nil {
		return false, err
	}

	err = authz.Authorized(ctx, r.system, team, authz.AccessLevelUpdate, authz.ResourceTeams)
	if err != nil {
		return false, err
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Create(corr).Error
	if err != nil {
		return false, fmt.Errorf("unable to create correlation for audit log")
	}

	r.auditLogger.Logf(console_reconciler.OpSyncTeam, *corr, *r.system, authz.UserFromContext(ctx), team, nil, "Manual sync requested")

	team, err = r.teamWithAssociations(*team.ID)
	if err != nil {
		return false, fmt.Errorf("unable to fetch team: %w", err)
	}

	r.teamReconciler <- reconcilers.Input{
		Corr: *corr,
		Team: *team,
	}

	return true, nil
}

func (r *queryResolver) Teams(ctx context.Context, pagination *model.Pagination, query *model.TeamsQuery, sort *model.TeamsSort) (*model.Teams, error) {
	err := authz.Authorized(ctx, r.system, nil, authz.AccessLevelRead, authz.ResourceTeams)
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
	pageInfo, err := r.paginatedQuery(pagination, query, sort, &dbmodels.Team{}, &teams)
	return &model.Teams{
		PageInfo: pageInfo,
		Nodes:    teams,
	}, err
}

func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*dbmodels.Team, error) {
	team := &dbmodels.Team{}
	err := r.db.Where("id = ?", id).First(team).Error
	if err != nil {
		return nil, err
	}

	err = authz.Authorized(ctx, r.system, team, authz.AccessLevelRead, authz.ResourceTeams)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *teamResolver) Users(ctx context.Context, obj *dbmodels.Team) ([]*dbmodels.User, error) {
	users := make([]*dbmodels.User, 0)
	err := r.db.Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *teamResolver) Metadata(ctx context.Context, obj *dbmodels.Team) (map[string]interface{}, error) {
	metadata := make([]*dbmodels.TeamMetadata, 0)
	err := r.db.Model(obj).Association("Metadata").Find(&metadata)
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
	err := r.db.Model(obj).Association("AuditLogs").Find(&auditLogs)
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
