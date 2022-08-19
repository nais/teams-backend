package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsCreate)
	if err != nil {
		return nil, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}

	teamUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	team, err := qtx.CreateTeam(ctx, sqlc.CreateTeamParams{
		ID:          teamUUID,
		Slug:        string(*input.Slug),
		Name:        input.Name,
		Purpose:     sql.NullString{String: *input.Purpose},
		CreatedByID: uuid.NullUUID{UUID: *actor.ID},
	})
	if err != nil {
		return nil, err
	}

	userTeamUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	err = qtx.AddUserToTeam(ctx, sqlc.AddUserToTeamParams{
		ID:          userTeamUUID,
		UserID:      *actor.ID,
		TeamID:      team.ID,
		CreatedByID: uuid.NullUUID{UUID: *actor.ID},
	})
	if err != nil {
		return nil, err
	}

	role, err := qtx.GetRoleByName(ctx, roles.RoleTeamOwner)
	if err != nil {
		return nil, err
	}

	userRoleUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	err = qtx.AddRoleToUser(ctx, sqlc.AddRoleToUserParams{
		ID:          userRoleUUID,
		UserID:      *actor.ID,
		RoleID:      role.ID,
		CreatedByID: uuid.NullUUID{UUID: *actor.ID},
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpCreateTeam, *corr, r.system, actor, team, nil, "Team created")
	i, err := r.newReconcilerInput(ctx, team.ID, *corr)
	if err != nil {
		return nil, err
	}

	r.teamReconciler <- *i
	return team, nil
}

func (r *mutationResolver) UpdateTeam(ctx context.Context, teamID *uuid.UUID, input model.UpdateTeamInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", teamID).First(team).Error
	if err != nil {
		return nil, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}

	err = r.gorm.Transaction(func(tx *gorm.DB) error {
		if input.Name != nil {
			team.Name = *input.Name
		}

		if input.Purpose != nil {
			team.Purpose = input.Purpose
		}

		err = db.UpdateTrackedObject(ctx, tx, team)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpUpdateTeam, *corr, r.system, actor, team, nil, "Team updated")

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

func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, input model.RemoveUsersFromTeamInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToRemove := make([]*dbmodels.User, 0)
	err = r.gorm.Where("id IN (?)", input.UserIds).Find(&usersToRemove).Error
	if err != nil {
		return nil, err
	}

	if len(usersToRemove) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}

	err = r.gorm.Transaction(func(tx *gorm.DB) error {
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
		r.auditLogger.Logf(console_reconciler.OpRemoveTeamMember, *corr, r.system, actor, team, removedUser, "Removed user '%s' from team '%s'", removedUser.Name, team.Name)
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

func (r *mutationResolver) SynchronizeTeam(ctx context.Context, teamID *uuid.UUID) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", teamID).First(team).Error
	if err != nil {
		return nil, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpSyncTeam, *corr, r.system, actor, team, nil, "Manual sync requested")

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

func (r *mutationResolver) AddTeamMembers(ctx context.Context, input model.AddTeamMembersInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToAdd := make([]*dbmodels.User, 0)
	err = r.gorm.Where("id IN (?)", input.UserIds).Find(&usersToAdd).Error
	if err != nil {
		return nil, err
	}

	if len(usersToAdd) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}
	err = r.gorm.Transaction(func(tx *gorm.DB) error {
		teamMemberRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamMember).First(teamMemberRole).Error
		if err != nil {
			return err
		}

		for _, user := range usersToAdd {
			isOwner, err := db.UserIsTeamOwner(tx, *user.ID, *team.ID)
			if err != nil {
				return err
			}

			if !isOwner {
				err = db.CreateTrackedObjectIgnoringDuplicates(ctx, tx, &dbmodels.UserRole{
					UserID:   *user.ID,
					RoleID:   *teamMemberRole.ID,
					TargetID: team.ID,
				})
				if err != nil {
					return err
				}
			}

			err = db.CreateTrackedObjectIgnoringDuplicates(ctx, tx, &dbmodels.UserTeam{
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
		r.auditLogger.Logf(console_reconciler.OpAddTeamMember, *corr, r.system, actor, team, addedUser, "Added user '%s' to team '%s' as a member", addedUser.Name, team.Name)
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

func (r *mutationResolver) AddTeamOwners(ctx context.Context, input model.AddTeamOwnersInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	usersToAdd := make([]*dbmodels.User, 0)
	err = r.gorm.Where("id IN (?)", input.UserIds).Find(&usersToAdd).Error
	if err != nil {
		return nil, err
	}

	if len(usersToAdd) != len(input.UserIds) {
		return nil, fmt.Errorf("one or more non-existing or duplicate user IDs given as parameter")
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}
	err = r.gorm.Transaction(func(tx *gorm.DB) error {
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
			err = db.CreateTrackedObjectIgnoringDuplicates(ctx, tx, &dbmodels.UserRole{
				UserID:   *user.ID,
				RoleID:   *teamOwnerRole.ID,
				TargetID: team.ID,
			})
			if err != nil {
				return err
			}

			err = db.CreateTrackedObjectIgnoringDuplicates(ctx, tx, &dbmodels.UserTeam{
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
		r.auditLogger.Logf(console_reconciler.OpAddTeamOwner, *corr, r.system, actor, team, addedUser, "Added user '%s' to team '%s' as owner", addedUser.Name, team.Name)
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

func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, input model.SetTeamMemberRoleInput) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", input.TeamID).First(team).Error
	if err != nil {
		return nil, err
	}

	user := &dbmodels.User{}
	err = r.gorm.Where("id = ?", input.UserID).First(user).Error
	if err != nil {
		return nil, err
	}

	userTeam := &dbmodels.UserTeam{}
	err = r.gorm.Where("team_id = ? AND user_id = ?", team.ID, user.ID).First(userTeam).Error
	if err != nil {
		return nil, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}
	err = r.gorm.Transaction(func(tx *gorm.DB) error {
		teamMemberRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamMember).First(teamMemberRole).Error
		if err != nil {
			return err
		}

		teamOwnerRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleTeamOwner).First(teamOwnerRole).Error
		if err != nil {
			return err
		}

		err = tx.Where("role_id IN (?) AND user_id = ? AND target_id = ?", []*uuid.UUID{teamMemberRole.ID, teamOwnerRole.ID}, user.ID, team.ID).Delete(&dbmodels.UserRole{}).Error
		if err != nil {
			return err
		}

		userRole := &dbmodels.UserRole{
			UserID:   *user.ID,
			RoleID:   *teamMemberRole.ID,
			TargetID: team.ID,
		}

		if input.Role == model.TeamRoleOwner {
			userRole.RoleID = *teamOwnerRole.ID
		}

		err = db.CreateTrackedObject(ctx, tx, userRole)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpSetTeamMemberRole, *corr, r.system, actor, team, user, "Set team member role for '%s' to '%s' in team '%s'", user.Email, input.Role, team.Name)

	team, err = r.teamWithAssociations(*team.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch team: %w", err)
	}

	return team, nil
}

func (r *queryResolver) Teams(ctx context.Context, pagination *model.Pagination, query *model.TeamsQuery, sort *model.TeamsSort) (*model.Teams, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsList)
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

func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*sqlc.Team, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsRead, *id)
	if err != nil {
		return nil, err
	}

	team := &dbmodels.Team{}
	err = r.gorm.Where("id = ?", id).First(team).Error
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *teamResolver) Slug(ctx context.Context, obj *sqlc.Team) (*slug.Slug, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *teamResolver) Purpose(ctx context.Context, obj *sqlc.Team) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *teamResolver) Metadata(ctx context.Context, obj *sqlc.Team) (map[string]interface{}, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationTeamsRead, *obj.ID)
	if err != nil {
		return nil, err
	}

	metadata := make([]*dbmodels.TeamMetadata, 0)
	err = r.gorm.Model(obj).Association("Metadata").Find(&metadata)
	if err != nil {
		return nil, err
	}

	kv := make(map[string]interface{})

	for _, pair := range metadata {
		kv[pair.Key] = pair.Value
	}

	return kv, nil
}

func (r *teamResolver) AuditLogs(ctx context.Context, obj *sqlc.Team) ([]*dbmodels.AuditLog, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationAuditLogsRead, *obj.ID)
	if err != nil {
		return nil, err
	}

	auditLogs := make([]*dbmodels.AuditLog, 0)
	err = r.gorm.Model(obj).Association("AuditLogs").Find(&auditLogs)
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (r *teamResolver) Members(ctx context.Context, obj *sqlc.Team) ([]*model.TeamMember, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	users := make([]*dbmodels.User, 0)
	err = r.gorm.Model(obj).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}

	members := make([]*model.TeamMember, len(users))
	for idx, user := range users {
		role := model.TeamRoleMember
		isOwner, err := db.UserIsTeamOwner(r.gorm, *user.ID, *obj.ID)
		if err != nil {
			return nil, err
		}

		if isOwner {
			role = model.TeamRoleOwner
		}

		members[idx] = &model.TeamMember{
			User: user,
			Role: role,
		}
	}
	return members, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }

func (r *mutationResolver) newReconcilerInput(ctx context.Context, id uuid.UUID, corr sqlc.Correlation) (*reconcilers.Input, error) {
	var err error
	team, err := r.queries.GetTeam(ctx, id)
	if err != nil {
		return nil, err
	}

	members, err := r.queries.GetTeamMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	metadata, err := r.queries.GetTeamMetadata(ctx, id)
	if err != nil {
		return nil, err
	}

	return &reconcilers.Input{
		Corr:     corr,
		Team:     team,
		Members:  members,
		Metadata: metadata,
	}, nil
}
