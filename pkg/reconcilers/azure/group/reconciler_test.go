package azure_group_reconciler_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	domain := "example.com"
	teamName := "myteam"
	teamSlug := "slug"
	teamPurpose := sql.NullString{}
	teamPurpose.Scan("My purpose")

	ctx := context.Background()

	creds := clientcredentials.Config{}

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
		User: &sqlc.User{
			Email: "add@example.com",
		},
	}
	keepUser := &db.User{
		User: &sqlc.User{
			Email: "keeper@example.com",
		},
	}
	removeUser := &db.User{
		User: &sqlc.User{
			Email: "removemember@example.com",
		},
	}
	correlationID := uuid.New()
	team := db.Team{
		Team: &sqlc.Team{
			ID:      uuid.New(),
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: teamPurpose,
		},
		Members: []*db.User{addUser, keepUser},
	}

	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	systemName := azure_group_reconciler.Name
	logHook := test.NewGlobal()

	t.Run("happy case", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeMember.Mail).
			Return(removeUser, nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.MatchedBy(func(purpose *string) bool {
				return *purpose == teamPurpose.String
			})).
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
			On("Logf", ctx, sqlc.AuditActionAzureGroupCreate, correlationID, systemName, mock.Anything, &teamSlug, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, sqlc.AuditActionAzureGroupDeleteMember, correlationID, systemName, mock.Anything, &teamSlug, &removeMember.Mail, mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, sqlc.AuditActionAzureGroupAddMember, correlationID, systemName, mock.Anything, &teamSlug, &addUser.Email, mock.Anything, addUser.Email, group.MailNickname).
			Return(nil).
			Once()

		err := reconciler.Reconcile(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.MatchedBy(func(purpose *string) bool {
				return *purpose == teamPurpose.String
			})).
			Return(nil, false, fmt.Errorf("GetOrCreateGroup failed")).
			Once()

		err := reconciler.Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.MatchedBy(func(purpose *string) bool {
				return *purpose == teamPurpose.String
			})).
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
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		team := db.Team{
			Team: &sqlc.Team{
				ID:      uuid.New(),
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: teamPurpose,
			},
		}

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.MatchedBy(func(purpose *string) bool {
				return *purpose == teamPurpose.String
			})).
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
		assert.Contains(t, logHook.LastEntry().Message, "unable to remove member 'removemember@example.com' from group 'nais-team-myteam'")
	})

	t.Run("GetUser fail", func(t *testing.T) {
		logHook.Reset()
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeMember.Mail).
			Return(removeUser, nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.Anything).
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
			On("Logf", ctx, sqlc.AuditActionAzureGroupDeleteMember, correlationID, systemName, mock.Anything, &teamSlug, &removeMember.Mail, mock.Anything, removeMember.Mail, group.MailNickname).
			Return(nil).
			Once()

		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(logHook.Entries))
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, "unable to lookup user with email 'add@example.com' in Azure")
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {
		logHook.Reset()
		database := db.NewMockDatabase(t)
		mockClient := azureclient.NewMockClient(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := azure_group_reconciler.New(database, auditLogger, creds, mockClient, domain)

		database.
			On("LoadSystemState", ctx, azure_group_reconciler.Name, team.ID, mock.Anything).
			Return(nil).
			Once()

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, mock.Anything).
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
		assert.Contains(t, logHook.LastEntry().Message, "unable to add member 'add@example.com' to Azure group 'nais-team-myteam'")
	})
}
