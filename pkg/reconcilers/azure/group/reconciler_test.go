package azure_group_reconciler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/mock"

	"github.com/nais/console/pkg/auditlogger"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
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
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, log)

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
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 1 && t[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupCreate && f.CorrelationID == correlationID
			}), mock.Anything, group.MailNickname, group.ID).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == removeMember.Mail
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == addUser.Email
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupAddMember && f.CorrelationID == correlationID
			}), mock.Anything, addUser.Email, group.MailNickname).
			Return(nil).
			Once()

		err := reconciler.Reconcile(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, log)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamPurpose).
			Return(nil, false, fmt.Errorf("GetOrCreateGroup failed")).
			Once()

		err := reconciler.Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, log)

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

		err := reconciler.Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("RemoveMemberFromGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)

		removeMemberFromGroupErr := errors.New("RemoveMemberFromGroup failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithError", removeMemberFromGroupErr).Return(log.WithError(err)).Once()

		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, mockLogger)

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

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			CorrelationID: correlationID,
			Team:          team,
		})
		assert.NoError(t, err)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)

		getUserError := errors.New("GetUser failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithError", getUserError).Return(log.WithError(getUserError)).Once()
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, mockLogger)

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
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return t[0].Identifier == string(teamSlug) && t[1].Identifier == removeMember.Mail
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()

		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		addMemberToGroupError := errors.New("AddMemberToGroup failed")
		mockLogger := logger.NewMockLogger(t)
		mockLogger.On("WithError", addMemberToGroupError).Return(log.WithError(addMemberToGroupError)).Once()
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain, mockLogger)

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

		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}
