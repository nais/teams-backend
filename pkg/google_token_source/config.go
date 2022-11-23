package google_token_source

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
)

func baseConfig(projectID string, scopes []string) impersonate.CredentialsConfig {
	return impersonate.CredentialsConfig{
		TargetPrincipal: serviceAccountName(projectID),
		Scopes:          scopes,
	}
}

func serviceAccountName(projectID string) string {
	return fmt.Sprintf("console@%s.iam.gserviceaccount.com", projectID)
}

func delegatedUserEmail(tenantDomain string) string {
	return "nais-console@" + tenantDomain
}

func GetDelegatedTokenSource(ctx context.Context, projectID, tenantDomain string, scopes []string) (oauth2.TokenSource, error) {
	cfg := baseConfig(projectID, scopes)
	cfg.Subject = delegatedUserEmail(tenantDomain)

	return impersonate.CredentialsTokenSource(ctx, cfg)
}

func GetTokenSource(ctx context.Context, projectID string, scopes []string) (oauth2.TokenSource, error) {
	cfg := baseConfig(projectID, scopes)

	return impersonate.CredentialsTokenSource(ctx, cfg)
}
