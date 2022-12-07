package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/gcp"
)

type GCP struct {
	Clusters       gcp.Clusters `envconfig:"CONSOLE_GCP_CLUSTERS"`
	CnrmRole       string       `envconfig:"CONSOLE_GCP_CNRM_ROLE"`
	BillingAccount string       `envconfig:"CONSOLE_GCP_BILLING_ACCOUNT"`
}

type NaisNamespace struct {
	AzureEnabled bool `envconfig:"CONSOLE_NAIS_NAMESPACE_AZURE_ENABLED"`
}

type UserSync struct {
	Enabled bool `envconfig:"CONSOLE_USERSYNC_ENABLED"`
}

type OAuth struct {
	ClientID     string `envconfig:"CONSOLE_OAUTH_CLIENT_ID"`
	ClientSecret string `envconfig:"CONSOLE_OAUTH_CLIENT_SECRET"`
	RedirectURL  string `envconfig:"CONSOLE_OAUTH_REDIRECT_URL"`
}

type NaisDeploy struct {
	Endpoint     string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT"`
	ProvisionKey string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`
}

type Config struct {
	GCP                       GCP
	UserSync                  UserSync
	NaisDeploy                NaisDeploy
	NaisNamespace             NaisNamespace
	OAuth                     OAuth
	GoogleManagementProjectID string `envconfig:"CONSOLE_GOOGLE_MANAGEMENT_PROJECT_ID"`
	TenantName                string `envconfig:"CONSOLE_TENANT_NAME"`
	TenantDomain              string `envconfig:"CONSOLE_TENANT_DOMAIN"`
	FrontendURL               string `envconfig:"CONSOLE_FRONTEND_URL"`
	DatabaseURL               string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress             string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
	LogFormat                 string `envconfig:"CONSOLE_LOG_FORMAT"`
	LogLevel                  string `envconfig:"CONSOLE_LOG_LEVEL"`
	AdminApiKey               string `envconfig:"CONSOLE_ADMIN_API_KEY"`
	StaticServiceAccounts     string `envconfig:"CONSOLE_STATIC_SERVICE_ACCOUNTS"`
}

type ImporterConfig struct {
	AzureClientID     string       `envconfig:"CONSOLE_IMPORTER_AZURE_CLIENT_ID"`
	AzureClientSecret string       `envconfig:"CONSOLE_IMPORTER_AZURE_CLIENT_SECRET"`
	AzureTenantID     string       `envconfig:"CONSOLE_IMPORTER_AZURE_TENANT_ID"`
	DatabaseURL       string       `envconfig:"CONSOLE_DATABASE_URL"`
	TenantDomain      string       `envconfig:"CONSOLE_TENANT_DOMAIN"`
	GCPClusters       gcp.Clusters `envconfig:"CONSOLE_GCP_CLUSTERS"`
}

func Defaults() *Config {
	return &Config{
		DatabaseURL:   "postgres://console:console@localhost:3002/console?sslmode=disable",
		FrontendURL:   "http://localhost:3001",
		ListenAddress: "127.0.0.1:3000",
		TenantName:    "example",
		TenantDomain:  "example.com",
		LogFormat:     "text",
		LogLevel:      "DEBUG",
		NaisDeploy: NaisDeploy{
			Endpoint: "http://localhost:8080/api/v1/provision",
		},
	}
}

func New() (*Config, error) {
	cfg := Defaults()

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewImporterConfig() (*ImporterConfig, error) {
	cfg := &ImporterConfig{
		DatabaseURL:  "postgres://console:console@localhost:3002/console?sslmode=disable",
		TenantDomain: "example.com",
	}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
