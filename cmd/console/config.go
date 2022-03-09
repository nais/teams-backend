package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	DatabaseURL             string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress           string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
	GoogleDomain            string `envconfig:"CONSOLE_GOOGLE_DOMAIN"`
	GoogleDelegatedUser     string `envconfig:"CONSOLE_GOOGLE_DELEGATED_USER"`
	GoogleCredentialsFile   string `envconfig:"CONSOLE_GOOGLE_CREDENTIALS_FILE"`
	GitHubOrganization      string `envconfig:"GITHUB_ORGANIZATION"`
	GitHubAppId             int64  `envconfig:"GITHUB_APP_ID"`
	GitHubAppInstallationId int64  `envconfig:"GITHUB_APP_INSTALLATION_ID"`
	GitHubPrivateKeyPath    string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`
}

func defaultconfig() *config {
	return &config{
		DatabaseURL:   "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		ListenAddress: "127.0.0.1:3000",
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
