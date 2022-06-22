package azure_group_reconciler_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/test"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const (
		domain   = "example.com"
		teamName = "myteam"
		teamSlug = "slug"
	)

	teamPurpose := console.Strp("My purpose")

	ctx := context.Background()

	hook := logrustest.NewGlobal()

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
		ID:   "some-keepMember-id",
		Mail: "keeper@example.com",
	}
	removeMember := &azureclient.Member{
		ID:   "some-removeMember-id",
		Mail: "removeMember@example.com",
	}
	addUser := &dbmodels.User{
		Email: "add@example.com",
	}
	keepUser := &dbmodels.User{
		Email: keepMember.Mail,
	}
	corr := dbmodels.Correlation{Model: modelWithId()}
	system := dbmodels.System{Model: modelWithId()}
	team := dbmodels.Team{
		Model:   modelWithId(),
		Slug:    teamSlug,
		Name:    teamName,
		Purpose: teamPurpose,
		Users: []*dbmodels.User{
			addUser, keepUser,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		auditLogger := auditlogger.New(db)

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(db, system, auditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
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

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			Corr: corr,
			Team: team,
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		auditLogger := auditlogger.New(db)

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(db, system, auditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
			Return(nil, false, fmt.Errorf("GetOrCreateGroup failed")).
			Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			Corr: corr,
			Team: team,
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		auditLogger := auditlogger.New(db)

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(db, system, auditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
			Return(group, true, nil).
			Once()
		mockClient.
			On("ListGroupMembers", mock.Anything, group).
			Return(nil, fmt.Errorf("ListGroupMembers failed")).
			Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			Corr: corr,
			Team: team,
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveMemberFromGroup fail", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})

		mockClient := &azureclient.MockClient{}
		mockAuditLogger := &auditlogger.MockAuditLogger{}
		reconciler := azure_group.New(db, system, mockAuditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
			Return(group, false /* pass false to skip the initial audit log entry */, nil).
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
			Corr: corr,
			Team: dbmodels.Team{
				Model:   modelWithId(),
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: teamPurpose,
			},
		})

		assert.NoError(t, err)
		mockAuditLogger.AssertNotCalled(t, "Logf")
		mockClient.AssertExpectations(t)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		auditLogger := auditlogger.New(db)

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(db, system, auditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
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
			Return(nil, fmt.Errorf("GetUser failed")).
			Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			Corr: corr,
			Team: team,
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})

		mockClient := &azureclient.MockClient{}
		mockAuditLogger := &auditlogger.MockAuditLogger{}
		reconciler := azure_group.New(db, system, mockAuditLogger, creds, mockClient, domain)

		mockClient.
			On("GetOrCreateGroup", mock.Anything, mock.Anything, "nais-team-slug", teamName, teamPurpose).
			Return(group, false /* pass false to skip the initial audit log entry */, nil).
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

		hook.Reset()
		err := reconciler.Reconcile(ctx, reconcilers.Input{
			Corr: corr,
			Team: team,
		})

		assert.NoError(t, err)
		mockAuditLogger.AssertNotCalled(t, "Logf")
		mockClient.AssertExpectations(t)
	})
}

func modelWithId() dbmodels.Model {
	id, _ := uuid.NewUUID()
	return dbmodels.Model{ID: &id}
}
