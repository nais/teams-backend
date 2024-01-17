package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/nais/teams-backend/pkg/fixtures"
	"github.com/nais/teams-backend/pkg/gcp"
)

type DependencyTrack struct {
	// Endpoint URL to the DependencyTrack API.
	Endpoint string `envconfig:"TEAMS_BACKEND_DEPENDENCYTRACK_ENDPOINT"`
	// Username The username to use when authenticating with DependencyTrack.
	Username string `envconfig:"TEAMS_BACKEND_DEPENDENCYTRACK_USERNAME"`
	// Password The password to use when authenticating with DependencyTrack.
	Password string `envconfig:"TEAMS_BACKEND_DEPENDENCYTRACK_PASSWORD"`
}

type GitHub struct {
	// Organization The GitHub organization slug for the tenant.
	Organization string `envconfig:"TEAMS_BACKEND_GITHUB_ORG"`

	// AuthEndpoint Endpoint URL to the GitHub auth component.
	AuthEndpoint string `envconfig:"TEAMS_BACKEND_GITHUB_AUTH_ENDPOINT"`
}

type GCP struct {
	// Clusters A JSON-encoded value describing the GCP clusters to use. Refer to the README for the format.
	Clusters gcp.Clusters `envconfig:"TEAMS_BACKEND_GCP_CLUSTERS"`

	// CnrmRole The name of the custom CNRM role that is used when creating role bindings for the GCP projects of each
	// team. The value must also contain the organization ID.
	//
	// Example: `organizations/<org_id>/roles/CustomCNRMRole`, where `<org_id>` is a numeric ID.
	CnrmRole string `envconfig:"TEAMS_BACKEND_GCP_CNRM_ROLE"`

	// BillingAccount The ID of the billing account that each team project will be assigned to.
	//
	// Example: `billingAccounts/123456789ABC`
	BillingAccount string `envconfig:"TEAMS_BACKEND_GCP_BILLING_ACCOUNT"`

	// WorkloadIdentityPoolName The name of the workload identity pool used in the management project.
	//
	// Example: projects/{project_number}/locations/global/workloadIdentityPools/{workload_identity_pool_id}
	WorkloadIdentityPoolName string `envconfig:"TEAMS_BACKEND_GCP_WORKLOAD_IDENTITY_POOL_NAME"`
}

type NaisNamespace struct {
	// AzureEnabled When set to true teams-backend will send the Azure group ID of the team, if it has been created by
	// the Azure AD group reconciler, to naisd when creating a namespace for the NAIS team.
	AzureEnabled bool `envconfig:"TEAMS_BACKEND_NAIS_NAMESPACE_AZURE_ENABLED"`
}

type UserSync struct {
	// Enabled When set to true teams-backend will keep the user database in sync with the connected Google
	// organization. The Google organization will be treated as the master.
	Enabled bool `envconfig:"TEAMS_BACKEND_USERSYNC_ENABLED"`

	// AdminGroupPrefix The prefix of the admin group email address.
	AdminGroupPrefix string `envconfig:"TEAMS_BACKEND_USERSYNC_ADMIN_GROUP_PREFIX" default:"console-admins"`

	// RunsToStore Number of runs to store for the userSync GraphQL query.
	RunsToStore int `envconfig:"TEAMS_BACKEND_USERSYNC_RUNS_TO_STORE" default:"5"`
}

type OAuth struct {
	// ClientID The ID of the OAuth 2.0 client to use for the OAuth login flow.
	ClientID string `envconfig:"TEAMS_BACKEND_OAUTH_CLIENT_ID"`

	// ClientSecret The client secret to use for the OAuth login flow.
	ClientSecret string `envconfig:"TEAMS_BACKEND_OAUTH_CLIENT_SECRET"`

	// RedirectURL The URL that Google will redirect back to after performing authentication.
	RedirectURL string `envconfig:"TEAMS_BACKEND_OAUTH_REDIRECT_URL"`
}

type NaisDeploy struct {
	// Endpoint URL to the NAIS deploy key provisioning endpoint
	Endpoint string `envconfig:"TEAMS_BACKEND_NAIS_DEPLOY_ENDPOINT" default:"http://localhost:8080/api/v1/provision"`

	// ProvisionKey The API key used when provisioning deploy keys on behalf of NAIS teams.
	ProvisionKey string `envconfig:"TEAMS_BACKEND_NAIS_DEPLOY_PROVISION_KEY"`

	// DeployKeyEndpoint URL to the NAIS deploy key endpoint
	DeployKeyEndpoint string `envconfig:"TEAMS_BACKEND_NAIS_DEPLOY_DEPLOY_KEY_ENDPOINT" default:"http://localhost:8080/internal/api/v1/apikey"`
}

type IAP struct {
	// IAP audience for validating IAP tokens
	Audience string `envconfig:"TEAMS_BACKEND_IAP_AUDIENCE"`
	// Insecure bypasses IAP authentication, just using the email header
	Insecure bool `envconfig:"TEAMS_BACKEND_IAP_INSECURE"`
}

type Config struct {
	DependencyTrack DependencyTrack
	GitHub          GitHub
	GCP             GCP
	UserSync        UserSync
	NaisDeploy      NaisDeploy
	NaisNamespace   NaisNamespace
	OAuth           OAuth
	IAP             IAP

	// Environments A list of environment names used for instance in GCP
	Environments []string

	// IgnoredEnvironments list of environments that won't be reconciled
	IgnoredEnvironments []string `envconfig:"TEAMS_BACKEND_IGNORED_ENVIRONMENTS"`

	// DatabaseURL The URL for the database.
	DatabaseURL string `envconfig:"TEAMS_BACKEND_DATABASE_URL" default:"postgres://console:console@localhost:3002/console?sslmode=disable"`

	// FrontendURL URL to the teams-frontend instance.
	FrontendURL string `envconfig:"TEAMS_BACKEND_FRONTEND_URL" default:"http://localhost:3001"`

	// Names of reconcilers to enable on first run of teams-backend
	//
	// Example: google:gcp:project,nais:namespace
	// Valid: [google:gcp:project|google:workspace-admin|nais:namespace|nais:deploy]
	FirstRunEnableReconcilers []fixtures.EnableableReconciler `envconfig:"TEAMS_BACKEND_FIRST_RUN_ENABLE_RECONCILERS"`

	// ListenAddress The host:port combination used by the http server.
	ListenAddress string `envconfig:"TEAMS_BACKEND_LISTEN_ADDRESS" default:"127.0.0.1:3000"`

	// LogFormat Customize the log format. Can be "text" or "json".
	LogFormat string `envconfig:"TEAMS_BACKEND_LOG_FORMAT" default:"text"`

	// LogLevel The log level used in teams-backend.
	LogLevel string `envconfig:"TEAMS_BACKEND_LOG_LEVEL" default:"DEBUG"`

	// GoogleManagementProjectID The ID of the NAIS management project in the tenant organization in GCP.
	GoogleManagementProjectID string `envconfig:"TEAMS_BACKEND_GOOGLE_MANAGEMENT_PROJECT_ID"`

	// OnpremClusters a list of onprem clusters (NAV only)
	// Example: "dev-fss,prod-fss,ci-fss"
	OnpremClusters []string `envconfig:"TEAMS_BACKEND_ONPREM_CLUSTERS"`

	// StaticServiceAccounts A JSON-encoded value describing a set of service accounts to be created when the
	// application starts. Refer to the README for the format.
	StaticServiceAccounts fixtures.ServiceAccounts `envconfig:"TEAMS_BACKEND_STATIC_SERVICE_ACCOUNTS"`

	// TenantDomain The domain for the tenant.
	TenantDomain string `envconfig:"TEAMS_BACKEND_TENANT_DOMAIN" default:"example.com"`

	// TenantName The name of the tenant.
	TenantName string `envconfig:"TEAMS_BACKEND_TENANT_NAME" default:"example"`
}

func New() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	cfg.ParseEnvironments()

	return cfg, nil
}

func (cfg *Config) ParseEnvironments() {
	var gcpEnvironments []string
	gcpClusters := make(map[string]gcp.Cluster)
	for environment, cluster := range cfg.GCP.Clusters {
		if !contains(cfg.IgnoredEnvironments, environment) {
			gcpClusters[environment] = cluster
			gcpEnvironments = append(gcpEnvironments, environment)
		}
	}

	var onpremEnvironments []string
	for _, environment := range cfg.OnpremClusters {
		if !contains(cfg.IgnoredEnvironments, environment) {
			onpremEnvironments = append(onpremEnvironments, environment)
		}
	}

	cfg.GCP.Clusters = gcpClusters
	cfg.OnpremClusters = onpremEnvironments
	cfg.Environments = append(gcpEnvironments, onpremEnvironments...)
}

func contains(haystack []string, needle string) bool {
	for _, element := range haystack {
		if strings.EqualFold(strings.TrimSpace(element), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}
