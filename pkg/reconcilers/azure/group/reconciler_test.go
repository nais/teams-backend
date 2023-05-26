package azure_group_reconciler_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/azureclient"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/stretchr/testify/mock"

	"github.com/nais/teams-backend/pkg/auditlogger"
	azure_group_reconciler "github.com/nais/teams-backend/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	domain := "example.com"
	teamSlug := slug.Slug("slug")
	teamPurpose := "My purpose"

	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	ctx := context.Background()

	group := &azureclient.Group{
		ID:           "some-group-id",
		MailNickname: "nais-team-myteam",
	}
	addMember := &azureclient.Member{
		ID:   "some-addMember-id",
		Mail: "add@example.com",
	}
	keepMember := &azureclient.Member{
		ID:   "some-keepmember-id",
		Mail: "keeper@example.com",
	}
	removeMember := &azureclient.Member{
		ID:   "some-removeMember-id",
		Mail: "removemember@example.com",
	}
	addUser := &db.User{
		User: &sqlc.User{Email: "add@example.com"},
	}
	keepUser := &db.User{
		User: &sqlc.User{Email: "keeper@example.com"},
	}
	removeUser := &db.User{
		User: &sqlc.User{Email: "removemember@example.com"},
	}
	correlationID := uuid.New()
	team := db.Team{
		Team: &sqlc.Team{
			Slug:    teamSlug,
			Purpose: teamPurpose,
		},
	}

	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers:   []*db.User{addUser, keepUser},
	}

	t.Run("happy case", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeMember.Mail).
			Return(removeUser, nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamPurpose).
			Return(group, true, nil).
			Once()
		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).
			Once()
		mockClient.
			On("RemoveMemberFromGroup", mock.Anything, group, removeMember).
			Return(nil).
			Once()
		mockClient.
			On("GetUser", mock.Anything, addUser.Email).
			Return(addMember, nil).
			Once()
		mockClient.
			On("AddMemberToGroup", mock.Anything, group, addMember).
			Return(nil).
			Once()

		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 1 && t[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupCreate && f.CorrelationID == correlationID
			}), mock.Anything, group.MailNickname, group.ID).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == removeMember.Mail
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == addUser.Email
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupAddMember && f.CorrelationID == correlationID
			}), mock.Anything, addUser.Email, group.MailNickname).
			Return(nil).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, log).
			Reconcile(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamPurpose).
			Return(nil, false, fmt.Errorf("GetOrCreateGroup failed")).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, log).
			Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamPurpose).
			Return(group, false, nil).
			Once()
		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return(nil, fmt.Errorf("ListGroupMembers failed")).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, log).
			Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("RemoveMemberFromGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		removeMemberFromGroupErr := errors.New("RemoveMemberFromGroup failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).Return(mockLogger).Once()
		mockLogger.On("WithError", removeMemberFromGroupErr).Return(log.WithError(err)).Once()

		team := db.Team{
			Team: &sqlc.Team{
				Slug:    teamSlug,
				Purpose: teamPurpose,
			},
		}

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamPurpose).
			Return(group, false, nil).
			Once()

		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{removeMember}, nil).
			Once()
		mockClient.
			On("RemoveMemberFromGroup", mock.Anything, group, removeMember).
			Return(removeMemberFromGroupErr).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, mockLogger).
			Reconcile(ctx, reconcilers.Input{
				CorrelationID: correlationID,
				Team:          team,
			})
		assert.NoError(t, err)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		getUserError := errors.New("GetUser failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).Return(mockLogger).Once()
		mockLogger.On("WithError", getUserError).Return(log.WithError(getUserError)).Once()

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeMember.Mail).
			Return(removeUser, nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", mock.Anything).
			Return(group, false, nil).
			Once()
		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).
			Once()
		mockClient.
			On("RemoveMemberFromGroup", mock.Anything, group, removeMember).
			Return(nil).
			Once()
		mockClient.
			On("GetUser", mock.Anything, addUser.Email).
			Return(nil, getUserError).
			Once()

		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return t[0].Identifier == string(teamSlug) && t[1].Identifier == removeMember.Mail
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, mockLogger).
			Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		addMemberToGroupError := errors.New("AddMemberToGroup failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).Return(mockLogger).Once()
		mockLogger.On("WithError", addMemberToGroupError).Return(log.WithError(addMemberToGroupError)).Once()

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", mock.Anything).
			Return(group, false, nil).
			Once()
		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember}, nil).
			Once()
		mockClient.
			On("GetUser", mock.Anything, addUser.Email).
			Return(addMember, nil).
			Once()
		mockClient.
			On("AddMemberToGroup", mock.Anything, group, addMember).
			Return(addMemberToGroupError).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, mockClient, domain, mockLogger).
			Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}

func TestAzureReconciler_Delete(t *testing.T) {
	const tenantDomain = "example.com"

	correlationID := uuid.New()
	teamSlug := slug.Slug("slug")
	ctx := context.Background()
	log := logger.NewMockLogger(t)
	azureClient := azureclient.NewMockClient(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)

	t.Run("Unable to load state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, teamSlug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()

		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		log.
			On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).
			Return(log).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, azureClient, tenantDomain, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "load reconciler state")
	})

	t.Run("Empty state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, teamSlug, mock.Anything).
			Return(nil).
			Once()

		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		log.
			On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).
			Return(log).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, azureClient, tenantDomain, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "missing group ID in reconciler state")
	})

	t.Run("Azure client error", func(t *testing.T) {
		grpID := uuid.New()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.AzureState)
				state.GroupID = &grpID
			}).
			Return(nil).
			Once()

		azureClient := azureclient.NewMockClient(t)
		azureClient.
			On("DeleteGroup", ctx, grpID).
			Return(fmt.Errorf("some error")).
			Once()

		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()

		log.
			On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).
			Return(log).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, azureClient, tenantDomain, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "delete Azure AD group with ID")
	})

	t.Run("Successful delete", func(t *testing.T) {
		grpID := uuid.New()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.AzureState)
				state.GroupID = &grpID
			}).
			Return(nil).
			Once()
		database.
			On("RemoveReconcilerStateForTeam", ctx, azure_group_reconciler.Name, teamSlug).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithSystemName", sqlc.SystemNameAzureGroup).
			Return(auditLogger).
			Once()
		auditLogger.
			On(
				"Logf",
				ctx,
				database,
				mock.MatchedBy(func(targets []auditlogger.Target) bool {
					return targets[0].Type == sqlc.AuditLogsTargetTypeTeam && targets[0].Identifier == string(teamSlug)
				}),
				mock.MatchedBy(func(fields auditlogger.Fields) bool {
					return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionAzureGroupDelete
				}),
				mock.MatchedBy(func(msg string) bool {
					return strings.HasPrefix(msg, "Delete Azure AD group")
				}),
				grpID,
			).
			Return(nil).
			Once()

		azureClient := azureclient.NewMockClient(t)
		azureClient.
			On("DeleteGroup", ctx, grpID).
			Return(nil).
			Once()

		log.
			On("WithSystem", string(sqlc.ReconcilerNameAzureGroup)).
			Return(log).
			Once()

		err := azure_group_reconciler.
			New(database, auditLogger, azureClient, tenantDomain, log).
			Delete(ctx, teamSlug, correlationID)
		assert.Nil(t, err)
	})
}
