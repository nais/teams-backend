package config

import (
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/legacy/envmap"
)

type GCP struct {
	// Clusters A JSON-encoded value describing the GCP clusters to use. Refer to the README for the format.
	Clusters gcp.Clusters `envconfig:"CONSOLE_GCP_CLUSTERS"`

	// CnrmRole The name of the custom CNRM role that is used when creating role bindings for the GCP projects of each
	// team. The value must also contain the organization ID.
	//
	// Example: `organizations/<org_id>/roles/CustomCNRMRole`, where `<org_id>` is a numeric ID.
	CnrmRole string `envconfig:"CONSOLE_GCP_CNRM_ROLE"`

	// BillingAccount The ID of the billing account that each team project will be assigned to.
	//
	// Example: `billingAccounts/123456789ABC`
	BillingAccount string `envconfig:"CONSOLE_GCP_BILLING_ACCOUNT"`
}

type NaisNamespace struct {
	// AzureEnabled When set to true Console will send the Azure group ID of the team, if it has been created by the
	// Azure AD group reconciler, to naisd when creating a namespace for the Console team.
	AzureEnabled bool `envconfig:"CONSOLE_NAIS_NAMESPACE_AZURE_ENABLED"`
}

type UserSync struct {
	// Enabled When set to true Console will keep the user database in sync with the connected Google organization. The
	// Google organization will be treated as the master.
	Enabled bool `envconfig:"CONSOLE_USERSYNC_ENABLED"`

	// AdminGroupPrefix The prefix of the admin group email address. Defaults to "console-admins".
	AdminGroupPrefix string `envconfig:"CONSOLE_USERSYNC_ADMIN_GROUP_PREFIX"`
}

type OAuth struct {
	// ClientID The ID of the OAuth 2.0 client to use for the OAuth login flow.
	ClientID string `envconfig:"CONSOLE_OAUTH_CLIENT_ID"`

	// ClientSecret The client secret to use for the OAuth login flow.
	ClientSecret string `envconfig:"CONSOLE_OAUTH_CLIENT_SECRET"`

	// RedirectURL The URL that Google will redirect back to after performing authentication.
	RedirectURL string `envconfig:"CONSOLE_OAUTH_REDIRECT_URL"`
}

type NaisDeploy struct {
	// Endpoint URL to the NAIS deploy key provisioning endpoint
	//
	// Example: `http://localhost:8080/api/v1/provision`
	Endpoint string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT"`

	// ProvisionKey The API key used when provisioning deploy keys on behalf of NAIS teams.
	ProvisionKey string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`

	// DeployKeyEndpoint URL to the NAIS deploy key endpoint
	//
	// Example: `http://localhost:8080/internal/api/v1/apikey`
	DeployKeyEndpoint string `envconfig:"CONSOLE_NAIS_DEPLOY_DEPLOY_KEY_ENDPOINT"`
}

type Config struct {
	GCP           GCP
	UserSync      UserSync
	NaisDeploy    NaisDeploy
	NaisNamespace NaisNamespace
	OAuth         OAuth

	// Environments A list of environment names used for instance in GCP
	Environments []string

	// DatabaseURL The URL for the database.
	//
	// Example: `postgres://console:console@localhost:3002/console?sslmode=disable`
	DatabaseURL string `envconfig:"CONSOLE_DATABASE_URL"`

	// FrontendURL URL to the console frontend.
	//
	// Example: `http://localhost:3001`
	FrontendURL string `envconfig:"CONSOLE_FRONTEND_URL"`

	// Names of reconcilers to enable on first run of console
	//
	// Example: google:gcp:project,nais:namespace
	// Valid: [google:gcp:project|google:workspace-admin|nais:namespace|nais:deploy]
	FirstRunEnableReconcilers []fixtures.EnableableReconciler `envconfig:"CONSOLE_FIRST_RUN_ENABLE_RECONCILERS"`

	// ListenAddress The host:port combination used by the http server.
	//
	// Example: `127.0.0.1:3000`
	ListenAddress string `envconfig:"CONSOLE_LISTEN_ADDRESS"`

	// LogFormat Customize the log format. Can be "text" or "json".
	LogFormat string `envconfig:"CONSOLE_LOG_FORMAT"`

	// LogLevel The log level used in console.
	LogLevel string `envconfig:"CONSOLE_LOG_LEVEL"`

	// GoogleManagementProjectID The ID of the NAIS management project in the tenant organization in GCP.
	GoogleManagementProjectID string `envconfig:"CONSOLE_GOOGLE_MANAGEMENT_PROJECT_ID"`

	// Maps an external Kubernetes cluster namespace onto permissions in a specific GCP project
	// Example: "dev-fss:dev prod-fss:rod dev-gcp:dev prod-gcp:prod"
	LegacyNaisNamespaces []envmap.EnvironmentMapping `envconfig:"CONSOLE_LEGACY_NAIS_NAMESPACES"`

	// Legacy cluster mapping. env:project
	// example: dev-gcp:nais-dev-123,prod-gcp:nais-prod-432
	LegacyClusters map[string]string

	// StaticServiceAccounts A JSON-encoded value describing a set of service accounts to be created when the
	// application starts. Refer to the README for the format.
	StaticServiceAccounts fixtures.ServiceAccounts `envconfig:"CONSOLE_STATIC_SERVICE_ACCOUNTS"`

	// TenantDomain The domain for the tenant.
	//
	// Example: `nav.no`
	TenantDomain string `envconfig:"CONSOLE_TENANT_DOMAIN"`

	// TenantName The name of the tenant.
	//
	// Example: `nav`.
	TenantName string `envconfig:"CONSOLE_TENANT_NAME"`

	// How long to wait until retrying a failed reconcile.
	ReconcileRetryInterval time.Duration `envconfig:"CONSOLE_RECONCILE_RETRY_INTERVAL"`
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
		LogFormat:     "text",
		LogLevel:      "DEBUG",
		TenantDomain:  "example.com",
		TenantName:    "example",
		NaisDeploy: NaisDeploy{
			Endpoint:          "http://localhost:8080/api/v1/provision",
			DeployKeyEndpoint: "http://localhost:8080/internal/api/v1/apikey",
		},
		ReconcileRetryInterval: time.Minute * 1,
		UserSync: UserSync{
			AdminGroupPrefix: "console-admins",
		},
	}
}

func New() (*Config, error) {
	cfg := Defaults()

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	environments := make([]string, 0)
	if strings.ToLower(cfg.TenantName) == "nav" {
		cfg.LegacyClusters = map[string]string{
			"dev-gcp":  "nais-dev-2e7b",
			"prod-gcp": "nais-prod-020f",
			"ci-gcp":   "nais-ci-e17f",
		}

		for _, mapping := range cfg.LegacyNaisNamespaces {
			environments = append(environments, mapping.Virtual)
		}

	} else {
		for environment := range cfg.GCP.Clusters {
			environments = append(environments, environment)
		}
	}

	cfg.Environments = environments

	return cfg, nil
}
