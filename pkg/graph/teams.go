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
	"gorm.io/gorm/clause"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*dbmodels.Team, error) {
	user := authz.UserFromContext(ctx)
	err := roles.RequireGlobalAuthorization(user, roles.AuthorizationTeamsCreate)
	if err != nil {
		return nil, err
	}

	corr := &dbmodels.Correlation{}
	team := &dbmodels.Team{
		Slug:    *input.Slug,
		Name:    input.Name,
		Purpose: input.Purpose,
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		err = r.createTrackedObject(ctx, team)
		if err != nil {
			return err
		}

		userTeam := &dbmodels.UserTeam{
			UserID: *user.ID,
			TeamID: *team.ID,
		}
		err = r.createTrackedObject(ctx, userTeam)
		if err != nil {
			return err
		}

		teamOwner := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamOwner).First(teamOwner).Error
		if err != nil {
			return err
		}

		userRole := &dbmodels.UserRole{
			UserID:   *user.ID,
			RoleID:   *teamOwner.ID,
			TargetID: team.ID,
		}
		err = tx.Create(userRole).Error
		if err != nil {
			return err
		}

		return nil
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
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.db.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToAdd := make([]*dbmodels.User, 0)
	err = r.db.Where("id IN (?)", input.UserIds).Find(&usersToAdd).Error
	if err != nil {
		return nil, err
	}

	if len(usersToAdd) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		teamMember := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamMember).First(teamMember).Error
		if err != nil {
			return err
		}

		for _, user := range usersToAdd {
			userRole := &dbmodels.UserRole{
				UserID:   *user.ID,
				RoleID:   *teamMember.ID,
				TargetID: team.ID,
			}
			err = tx.Create(userRole).Error
			if err != nil {
				return err
			}

			tm := &dbmodels.UserTeam{
				UserID: *user.ID,
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

	for _, addedUser := range usersToAdd {
		r.auditLogger.Logf(console_reconciler.OpAddTeamMember, *corr, *r.system, user, team, addedUser, "Added user '%s' to team '%s'", addedUser.Name, team.Name)
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
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.db.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToRemove := make([]*dbmodels.User, 0)
	err = r.db.Where("id IN (?)", input.UserIds).Find(&usersToRemove).Error
	if err != nil {
		return nil, err
	}

	if len(usersToRemove) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		err = tx.Where("user_id IN (?) AND target_id = ?", input.UserIds, team.ID).Delete(&dbmodels.UserRole{}).Error
		if err != nil {
			return err
		}

		err = tx.Where("user_id IN (?) AND team_id = ?", input.UserIds, team.ID).Delete(&dbmodels.UserTeam{}).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, removedUser := range usersToRemove {
		r.auditLogger.Logf(console_reconciler.OpRemoveTeamMember, *corr, *r.system, user, team, removedUser, "Removed user '%s' from team '%s'", removedUser.Name, team.Name)
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
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, *teamID)
	if err != nil {
		return false, err
	}

	team := &dbmodels.Team{}
	err = r.db.Where("id = ?", teamID).First(team).Error
	if err != nil {
		return false, err
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Create(corr).Error
	if err != nil {
		return false, fmt.Errorf("unable to create correlation for audit log")
	}

	r.auditLogger.Logf(console_reconciler.OpSyncTeam, *corr, *r.system, user, team, nil, "Manual sync requested")

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

func (r *mutationResolver) AddTeamOwners(ctx context.Context, input model.AddTeamOwnersInput) (*dbmodels.Team, error) {
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.db.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToAdd := make([]*dbmodels.User, 0)
	err = r.db.Where("id IN (?)", input.UserIds).Find(&usersToAdd).Error
	if err != nil {
		return nil, err
	}

	if len(usersToAdd) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr := &dbmodels.Correlation{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		// Remove the team member role that the user potentially has
		teamMemberRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamMember).First(teamMemberRole).Error
		if err != nil {
			return err
		}

		err = tx.Where("role_id = ? AND user_id IN (?) AND target_id = ?", teamMemberRole.ID, input.UserIds, team.ID).Delete(&dbmodels.UserRole{}).Error
		if err != nil {
			return err
		}

		teamOwnerRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamOwner).First(teamOwnerRole).Error
		if err != nil {
			return err
		}

		for _, user := range usersToAdd {
			// Ignore duplicate conflict that can occur if the user is already an owner of the team
			err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbmodels.UserRole{
				UserID:   *user.ID,
				RoleID:   *teamOwnerRole.ID,
				TargetID: team.ID,
			}).Error
			if err != nil {
				return err
			}

			err = r.createTrackedObjectIgnoringDuplicates(ctx, &dbmodels.UserTeam{
				UserID: *user.ID,
				TeamID: *team.ID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, addedUser := range usersToAdd {
		r.auditLogger.Logf(console_reconciler.OpAddTeamOwner, *corr, *r.system, user, team, addedUser, "Added user '%s' to team '%s' as owner", addedUser.Name, team.Name)
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

func (r *queryResolver) Teams(ctx context.Context, pagination *model.Pagination, query *model.TeamsQuery, sort *model.TeamsSort) (*model.Teams, error) {
	user := authz.UserFromContext(ctx)
	err := roles.RequireGlobalAuthorization(user, roles.AuthorizationTeamsList)
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
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsRead, *id)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.db.Where("id = ?", id).First(team).Error
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *teamResolver) Users(ctx context.Context, obj *dbmodels.Team) ([]*dbmodels.User, error) {
	user := authz.UserFromContext(ctx)
	err := roles.RequireGlobalAuthorization(user, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = r.db.Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *teamResolver) Metadata(ctx context.Context, obj *dbmodels.Team) (map[string]interface{}, error) {
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationTeamsRead, *obj.ID)
	if err != nil {
		return nil, err
	}

	metadata := make([]*dbmodels.TeamMetadata, 0)
	err = r.db.Model(obj).Association("Metadata").Find(&metadata)
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
	user := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(user, roles.AuthorizationAuditLogsRead, *obj.ID)
	if err != nil {
		return nil, err
	}

	auditLogs := make([]*dbmodels.AuditLog, 0)
	err = r.db.Model(obj).Association("AuditLogs").Find(&auditLogs)
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
