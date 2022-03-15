package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	DatabaseURL             string `envconfig:"CONSOLE_DATABASE_URL"`
	GitHubAppId             int64  `envconfig:"CONSOLE_GITHUB_APP_ID"`
	GitHubAppInstallationId int64  `envconfig:"CONSOLE_GITHUB_APP_INSTALLATION_ID"`
	GitHubOrganization      string `envconfig:"CONSOLE_GITHUB_ORGANIZATION"`
	GitHubPrivateKeyPath    string `envconfig:"CONSOLE_GITHUB_PRIVATE_KEY_PATH"`
	GoogleCredentialsFile   string `envconfig:"CONSOLE_GOOGLE_CREDENTIALS_FILE"`
	GoogleDelegatedUser     string `envconfig:"CONSOLE_GOOGLE_DELEGATED_USER"`
	GoogleDomain            string `envconfig:"CONSOLE_GOOGLE_DOMAIN"`
	ListenAddress           string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
	NaisDeployEndpoint      string `envconfig:"CONSOLE_NAIS_DEPLOY_ENDPOINT"`
	NaisDeployProvisionKey  string `envconfig:"CONSOLE_NAIS_DEPLOY_PROVISION_KEY"`
}

func defaultconfig() *config {
	return &config{
		DatabaseURL:        "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		ListenAddress:      "127.0.0.1:3000",
		NaisDeployEndpoint: "http://localhost:8080/api/v1/provision",
	}
}

func configure() (*config, error) {
	cfg := defaultconfig()

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
