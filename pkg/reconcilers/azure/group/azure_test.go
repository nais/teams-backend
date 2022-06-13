package azure_group_reconciler_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/test"
	"testing"
	"time"

	"github.com/nais/console/pkg/azureclient"
	"github.com/stretchr/testify/mock"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const domain = "example.com"
	const teamName = "myteam"
	purpose := helpers.Strp(teamName)
	teamSlug := dbmodels.Slug(teamName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	db := test.GetTestDB()
	auditLogger := auditlogger.New(db)

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
		Email: helpers.Strp("add@example.com"),
	}
	keepUser := &dbmodels.User{
		Email: &keepMember.Mail,
	}
	syncID, _ := uuid.NewUUID()
	sync := dbmodels.Synchronization{
		Model: dbmodels.Model{
			ID: &syncID,
		},
	}
	systemID, _ := uuid.NewUUID()
	system := dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(group, true, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(addMember, nil).Once()
		mockClient.On("AddMemberToGroup", mock.Anything, group, addMember).Return(nil).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(nil, false, fmt.Errorf("GetOrCreateGroup failed")).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(group, true, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return(nil, fmt.Errorf("ListGroupMembers failed")).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveMemberFromGroup fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(group, true, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).
			Return(fmt.Errorf("RemoveMemberFromGroup failed")).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(group, true, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(nil, fmt.Errorf("GetUser failed")).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(system, auditLogger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, purpose).
			Return(group, true, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(addMember, nil).Once()
		mockClient.On("AddMemberToGroup", mock.Anything, group, addMember).
			Return(fmt.Errorf("AddMemberToGroup failed")).Once()

		err := reconciler.Reconcile(ctx, sync, dbmodels.Team{
			Slug:    teamSlug,
			Name:    teamName,
			Purpose: purpose,
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
