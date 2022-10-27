package azure_group_reconciler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"

	"github.com/nais/console/pkg/auditlogger"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	domain := "example.com"
	teamSlug := slug.Slug("slug")
	teamPurpose := "My purpose"

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
			ID:      uuid.New(),
			Slug:    teamSlug,
			Purpose: teamPurpose,
		},
	}

	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers:   []*db.User{addUser, keepUser},
	}

	logHook := test.NewGlobal()

	t.Run("happy case", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
			}), mock.Anything, group).
			Return(nil)

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == removeMember.Mail
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil)

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(teamSlug) && t[1].Identifier == addUser.Email
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionAzureGroupAddMember && f.CorrelationID == correlationID
			}), mock.Anything, addUser.Email, group.MailNickname).
			Return(nil)

		err := reconciler.Reconcile(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
		logHook.Reset()
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		team := db.Team{
			Team: &sqlc.Team{
				ID:      uuid.New(),
				Slug:    teamSlug,
				Purpose: teamPurpose,
			},
		}

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
			Return(fmt.Errorf("RemoveMemberFromGroup failed")).
			Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			CorrelationID: correlationID,
			Team:          team,
		})
		assert.NoError(t, err)

		assert.Equal(t, 1, len(logHook.Entries))
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, `unable to remove member "removemember@example.com" from group "nais-team-myteam"`)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		logHook.Reset()
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
			Return(nil, fmt.Errorf("GetUser failed")).
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

		assert.Equal(t, 1, len(logHook.Entries))
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, `unable to lookup user with email "add@example.com" in Azure`)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {
		logHook.Reset()
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, mockClient, domain)

		database.
			On("LoadReconcilerStateForTeam", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
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
			Return(fmt.Errorf("AddMemberToGroup failed")).
			Once()

		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(logHook.Entries))
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, `unable to add member "add@example.com" to Azure group "nais-team-myteam"`)
	})
}
