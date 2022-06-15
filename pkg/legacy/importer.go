package legacy

import (
	"context"

	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

type GroupImporter struct {
	db     *gorm.DB
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

func (gimp *GroupImporter) GroupMembers(groupID string) ([]*dbmodels.User, error) {
	ctx := context.Background()
	members, err := gimp.client.ListGroupMembers(ctx, &azureclient.Group{
		ID: groupID,
	})
	if err != nil {
		return nil, err
	}
	users := make([]*dbmodels.User, 0, len(members))
	for _, member := range members {
		users = append(users, &dbmodels.User{
			Email: member.Mail,
			Name:  member.Mail,
		})
	}
	return users, nil
}

func (gimp *GroupImporter) GroupOwners(groupID string) ([]*dbmodels.User, error) {
	ctx := context.Background()
	members, err := gimp.client.ListGroupOwners(ctx, &azureclient.Group{
		ID: groupID,
	})
	if err != nil {
		return nil, err
	}
	users := make([]*dbmodels.User, 0, len(members))
	for _, member := range members {
		users = append(users, &dbmodels.User{
			Email: member.UserPrincipalName,
			Name:  member.UserPrincipalName,
		})
	}
	return users, nil
}
