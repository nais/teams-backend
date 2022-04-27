package azure_group_reconciler_test

import (
	"context"
	"github.com/nais/console/pkg/azureclient"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	azure_group "github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const teamName = "myteam"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

	creds := clientcredentials.Config{}
	mockClient := &azureclient.MockClient{}
	reconciler := azure_group.New(logger, creds, mockClient)

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
		Email: strp("add@example.com"),
	}
	keepUser := &dbmodels.User{
		Email: &keepMember.Mail,
	}

	mockClient.On("GetOrCreateGroup", mock.Anything, "nais-team-myteam", teamName, teamName).
		Return(group, nil)
	mockClient.On("ListGroupMembers", mock.Anything, group).
		Return([]*azureclient.Member{keepMember, removeMember}, nil)
	mockClient.On("RemoveMemberFromGroup", mock.Anything, group, removeMember).Return(nil)
	mockClient.On("GetUser", mock.Anything, *addUser.Email).Return(addMember, nil)
	mockClient.On("AddMemberToGroup", mock.Anything, group, addMember).Return(nil)

	err := reconciler.Reconcile(ctx, reconcilers.Input{
		System:          nil,
		Synchronization: nil,
		Team: &dbmodels.Team{
			Slug:    strp(teamName),
			Name:    strp(teamName),
			Purpose: strp(teamName),
			Users: []*dbmodels.User{
				addUser, keepUser,
			},
		},
	})

	assert.NoError(t, err)
}

func strp(s string) *string {
	return &s
}
