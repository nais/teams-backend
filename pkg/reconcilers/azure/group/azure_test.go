package azure_group_reconciler_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	helpers "github.com/nais/console/pkg/console"
	"testing"
	"time"

	"github.com/nais/console/pkg/azureclient"
	"github.com/stretchr/testify/mock"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const domain = "example.com"
	const teamName = "myteam"
	teamSlug := dbmodels.Slug(teamName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

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
	sync := &dbmodels.Synchronization{
		Model: dbmodels.Model{
			ID: &syncID,
		},
	}
	systemID, _ := uuid.NewUUID()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(group, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(addMember, nil).Once()
		mockClient.On("AddMemberToGroup", mock.Anything, group, addMember).Return(nil).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetOrCreateGroup fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(nil, fmt.Errorf("GetOrCreateGroup failed")).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListGroupMembers fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(group, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return(nil, fmt.Errorf("ListGroupMembers failed")).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveMemberFromGroup fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(group, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).
			Return(fmt.Errorf("RemoveMemberFromGroup failed")).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetUser fail", func(t *testing.T) {
		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(group, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(nil, fmt.Errorf("GetUser failed")).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AddMemberToGroup fail", func(t *testing.T) {

		mockClient := &azureclient.MockClient{}
		reconciler := azure_group.New(logger, creds, mockClient, domain)

		mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
			Return(group, nil).Once()
		mockClient.On("ListGroupMembers", mock.Anything, group).
			Return([]*azureclient.Member{keepMember, removeMember}, nil).Once()
		mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil).Once()
		mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(addMember, nil).Once()
		mockClient.On("AddMemberToGroup", mock.Anything, group, addMember).
			Return(fmt.Errorf("AddMemberToGroup failed")).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          *system,
			Synchronization: *sync,
			Team: &dbmodels.Team{
				Slug:    teamSlug,
				Name:    teamName,
				Purpose: helpers.Strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)

	})
}
