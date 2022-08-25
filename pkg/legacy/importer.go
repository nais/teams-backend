package legacy

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

type GroupImporter struct {
	oauth  clientcredentials.Config
	client azureclient.Client
}

func New(oauth clientcredentials.Config, client azureclient.Client) *GroupImporter {
	return &GroupImporter{
		oauth:  oauth,
		client: client,
	}
}

func NewFromConfig(cfg *config.Config) (*GroupImporter, error) {
	endpoint := microsoft.AzureADEndpoint(cfg.Azure.TenantID)
	conf := clientcredentials.Config{
		ClientID:     cfg.Azure.ClientID,
		ClientSecret: cfg.Azure.ClientSecret,
		TokenURL:     endpoint.TokenURL,
		AuthStyle:    endpoint.AuthStyle,
		Scopes: []string{
			"https://graph.microsoft.com/.default",
		},
	}

	return New(conf, azureclient.New(conf.Client(context.Background()))), nil
}

func (gimp *GroupImporter) GroupMembers(groupID string) ([]*db.User, error) {
	ctx := context.Background()
	members, err := gimp.client.ListGroupMembers(ctx, &azureclient.Group{
		ID: groupID,
	})
	if err != nil {
		return nil, err
	}
	return dbUsers(members), nil
}

func (gimp *GroupImporter) GroupOwners(groupID string) ([]*db.User, error) {
	ctx := context.Background()
	owners, err := gimp.client.ListGroupOwners(ctx, &azureclient.Group{
		ID: groupID,
	})
	if err != nil {
		return nil, err
	}
	return dbUsers(owners), nil
}

func dbUsers(members []*azureclient.Member) []*db.User {
	users := make([]*db.User, 0, len(members))
	for _, member := range members {
		users = append(users, &db.User{
			User: &sqlc.User{
				Email: member.Mail,
				Name:  member.Name(),
			},
		})
	}
	return users
}
