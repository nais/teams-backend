package legacy

import (
	"context"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

type Importer struct {
	oauth  clientcredentials.Config
	client azureclient.Client
}

func New(oauth clientcredentials.Config, client azureclient.Client) *Importer {
	return &Importer{
		oauth:  oauth,
		client: client,
	}
}

func NewFromConfig(cfg *config.Config, logger auditlogger.Logger) (*Importer, error) {
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
