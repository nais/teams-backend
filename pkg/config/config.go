package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/dependencytrack"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/gcp"
	"github.com/nais/console/pkg/legacy/envmap"
)

type DependencyTrack struct {
	// Instances A list of dependency track instances (one per cluster).
	Instances dependencytrack.Instances `envconfig:"CONSOLE_DEPENDENCYTRACK_INSTANCES"`
}

type GitHub struct {
	// Organization The GitHub organization slug for the tenant.
	Organization string `envconfig:"CONSOLE_GITHUB_ORG"`

	// AuthEndpoint Endpoint URL to the GitHub auth component.
	AuthEndpoint string `envconfig:"CONSOLE_GITHUB_AUTH_ENDPOINT"`
}

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

	// WorkloadIdentityPoolName The name of the workload identity pool used in the management project.
	//
	// Example: projects/{project_number}/locations/global/workloadIdentityPools/{workload_identity_pool_id}
	WorkloadIdentityPoolName string `envconfig:"CONSOLE_GCP_WORKLOAD_IDENTITY_POOL_NAME"`
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

	// AdminGroupPrefix The prefix of the admin group email address.
	AdminGroupPrefix string `envconfig:"CONSOLE_USERSYNC_ADMIN_GROUP_PREFIX" default:"console-admins"`

	// RunsToStore Number of runs to store for the userSync GraphQL query.
	RunsToStore int `envconfig:"CONSOLE_USERSYNC_RUNS_TO_STORE" default:"5"`
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
	Endpoint string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT" default:"http://localhost:8080/api/v1/provision"`

	// ProvisionKey The API key used when provisioning deploy keys on behalf of NAIS teams.
	ProvisionKey string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`

	// DeployKeyEndpoint URL to the NAIS deploy key endpoint
	DeployKeyEndpoint string `envconfig:"CONSOLE_NAIS_DEPLOY_DEPLOY_KEY_ENDPOINT" default:"http://localhost:8080/internal/api/v1/apikey"`
}

type Config struct {
	DependencyTrack DependencyTrack
	GitHub          GitHub
	GCP             GCP
	UserSync        UserSync
	NaisDeploy      NaisDeploy
	NaisNamespace   NaisNamespace
	OAuth           OAuth

	// Environments A list of environment names used for instance in GCP
	Environments []string

	// DatabaseURL The URL for the database.
	DatabaseURL string `envconfig:"CONSOLE_DATABASE_URL" default:"postgres://console:console@localhost:3002/console?sslmode=disable"`

	// FrontendURL URL to the console frontend.
	FrontendURL string `envconfig:"CONSOLE_FRONTEND_URL" default:"http://localhost:3001"`

	// Names of reconcilers to enable on first run of console
	//
	// Example: google:gcp:project,nais:namespace
	// Valid: [google:gcp:project|google:workspace-admin|nais:namespace|nais:deploy]
	FirstRunEnableReconcilers []fixtures.EnableableReconciler `envconfig:"CONSOLE_FIRST_RUN_ENABLE_RECONCILERS"`

	// ListenAddress The host:port combination used by the http server.
	ListenAddress string `envconfig:"CONSOLE_LISTEN_ADDRESS" default:"127.0.0.1:3000"`

	// LogFormat Customize the log format. Can be "text" or "json".
	LogFormat string `envconfig:"CONSOLE_LOG_FORMAT" default:"text"`

	// LogLevel The log level used in console.
	LogLevel string `envconfig:"CONSOLE_LOG_LEVEL" default:"DEBUG"`

	// GoogleManagementProjectID The ID of the NAIS management project in the tenant organization in GCP.
	GoogleManagementProjectID string `envconfig:"CONSOLE_GOOGLE_MANAGEMENT_PROJECT_ID"`

	// Maps an external Kubernetes cluster namespace onto permissions in a specific GCP project
	// Example: "dev-fss:dev prod-fss:prod dev-gcp:dev prod-gcp:prod"
	LegacyNaisNamespaces []envmap.EnvironmentMapping `envconfig:"CONSOLE_LEGACY_NAIS_NAMESPACES"`

	// StaticServiceAccounts A JSON-encoded value describing a set of service accounts to be created when the
	// application starts. Refer to the README for the format.
	StaticServiceAccounts fixtures.ServiceAccounts `envconfig:"CONSOLE_STATIC_SERVICE_ACCOUNTS"`

	// TenantDomain The domain for the tenant.
	TenantDomain string `envconfig:"CONSOLE_TENANT_DOMAIN" default:"example.com"`

	// TenantName The name of the tenant.
	TenantName string `envconfig:"CONSOLE_TENANT_NAME" default:"example"`
}

func New() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	environments := make([]string, 0)
	if strings.ToLower(cfg.TenantName) == "nav" {
		for _, mapping := range cfg.LegacyNaisNamespaces {
			environments = append(environments, mapping.Legacy)
		}
	} else {
		for environment := range cfg.GCP.Clusters {
			environments = append(environments, environment)
		}
	}

	cfg.Environments = environments

	return cfg, nil
}
