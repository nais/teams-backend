package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Azure reconciler
type Azure struct {
	Enabled      bool   `envconfig:"CONSOLE_AZURE_ENABLED"`
	ClientID     string `envconfig:"CONSOLE_AZURE_CLIENT_ID"`
	ClientSecret string `envconfig:"CONSOLE_AZURE_CLIENT_SECRET"`
	TenantID     string `envconfig:"CONSOLE_AZURE_TENANT_ID"`
}

// GitHub reconciler
type GitHub struct {
	Enabled           bool   `envconfig:"CONSOLE_GITHUB_ENABLED"`
	AppID             int64  `envconfig:"CONSOLE_GITHUB_APP_ID"`
	AppInstallationID int64  `envconfig:"CONSOLE_GITHUB_APP_INSTALLATION_ID"`
	Organization      string `envconfig:"CONSOLE_GITHUB_ORGANIZATION"`
	PrivateKeyPath    string `envconfig:"CONSOLE_GITHUB_PRIVATE_KEY_PATH"`
}

// Google workspace admin reconciler
type Google struct {
	Enabled         bool   `envconfig:"CONSOLE_GOOGLE_ENABLED"`
	DelegatedUser   string `envconfig:"CONSOLE_GOOGLE_DELEGATED_USER"`
	CredentialsFile string `envconfig:"CONSOLE_GOOGLE_CREDENTIALS_FILE"`
}

// Google Cloud Platform reconciler
type GCP struct {
	Enabled          bool              `envconfig:"CONSOLE_GCP_ENABLED"`
	ProjectParentIDs map[string]string `envconfig:"CONSOLE_GCP_PROJECT_PARENT_IDS"` // suffix is key, parentID is value
}

// Nais deploy reconciler
type NaisDeploy struct {
	Enabled      bool   `envconfig:"CONSOLE_NAIS_DEPLOY_ENABLED"`
	Endpoint     string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT"`
	ProvisionKey string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`
}

// Google Cloud Platform reconciler
type NaisNamespace struct {
	Enabled     bool   `envconfig:"CONSOLE_NAIS_NAMESPACE_ENABLED"`
	TopicPrefix string `envconfig:"CONSOLE_NAIS_NAMESPACE_TOPIC_PREFIX"`
	ProjectID   string `envconfig:"CONSOLE_NAIS_NAMESPACE_PROJECT_ID"`
}

// User synchronizer
type UserSync struct {
	Enabled bool `envconfig:"CONSOLE_USERSYNC_ENABLED"`
}

type OAuth struct {
	ClientID     string `envconfig:"CONSOLE_OAUTH_CLIENT_ID"`
	ClientSecret string `envconfig:"CONSOLE_OAUTH_CLIENT_SECRET"`
	RedirectURL  string `envconfig:"CONSOLE_OAUTH_REDIRECT_URL"`
}

type Config struct {
	Azure         Azure
	GitHub        GitHub
	Google        Google
	GCP           GCP
	UserSync      UserSync
	NaisDeploy    NaisDeploy
	NaisNamespace NaisNamespace
	OAuth         OAuth
	PartnerDomain string `envconfig:"CONSOLE_PARTNER_DOMAIN"`
	AutoLoginUser string `envconfig:"CONSOLE_AUTO_LOGIN_USER"`
	FrontendURL   string `envconfig:"CONSOLE_FRONTEND_URL"`
	DatabaseURL   string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
	LogFormat     string `envconfig:"CONSOLE_LOG_FORMAT"`
	LogLevel      string `envconfig:"CONSOLE_LOG_LEVEL"`
}

func Defaults() *Config {
	return &Config{
		DatabaseURL:   "postgres://console:console@localhost:5432/console?sslmode=disable",
		FrontendURL:   "http://localhost:3001",
		ListenAddress: "127.0.0.1:3000",
		NaisDeploy: NaisDeploy{
			Endpoint: "http://localhost:8080/api/v1/provision",
		},
		LogFormat: "text",
		LogLevel:  "DEBUG",
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
