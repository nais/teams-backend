package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	DatabaseURL           string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress         string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
	GoogleDomain          string `envconfig:"CONSOLE_GOOGLE_DOMAIN"`
	GoogleDelegatedUser   string `envconfig:"CONSOLE_GOOGLE_DELEGATED_USER"`
	GoogleCredentialsFile string `envconfig:"CONSOLE_GOOGLE_CREDENTIALS_FILE"`
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
