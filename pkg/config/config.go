package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Azure struct {
	Enabled      bool   `envconfig:"CONSOLE_AZURE_ENABLED"`
	ClientID     string `envconfig:"CONSOLE_AZURE_CLIENT_ID"`
	ClientSecret string `envconfig:"CONSOLE_AZURE_CLIENT_SECRET"`
	TenantID     string `envconfig:"CONSOLE_AZURE_TENANT_ID"`
	RedirectURL  string `envconfig:"CONSOLE_AZURE_REDIRECT_URL"`
}

type GitHub struct {
	Enabled           bool   `envconfig:"CONSOLE_GITHUB_ENABLED"`
	AppID             int64  `envconfig:"CONSOLE_GITHUB_APP_ID"`
	AppInstallationID int64  `envconfig:"CONSOLE_GITHUB_APP_INSTALLATION_ID"`
	Organization      string `envconfig:"CONSOLE_GITHUB_ORGANIZATION"`
	PrivateKeyPath    string `envconfig:"CONSOLE_GITHUB_PRIVATE_KEY_PATH"`
}

type Google struct {
	Enabled         bool   `envconfig:"CONSOLE_GOOGLE_ENABLED"`
	CredentialsFile string `envconfig:"CONSOLE_GOOGLE_CREDENTIALS_FILE"`
	DelegatedUser   string `envconfig:"CONSOLE_GOOGLE_DELEGATED_USER"`
	Domain          string `envconfig:"CONSOLE_GOOGLE_DOMAIN"`
}

type NaisDeploy struct {
	Enabled      bool   `envconfig:"CONSOLE_NAIS_DEPLOY_ENABLED"`
	Endpoint     string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT"`
	ProvisionKey string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`
}

type Config struct {
	Azure         Azure
	GitHub        GitHub
	Google        Google
	NaisDeploy    NaisDeploy
	DatabaseURL   string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
}

func Defaults() *Config {
	return &Config{
		DatabaseURL:   "postgres://console:console@localhost:5432/console?sslmode=disable",
		ListenAddress: "127.0.0.1:3000",
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
