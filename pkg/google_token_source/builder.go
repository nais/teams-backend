package google_token_source

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin_directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/impersonate"
)

type Builder struct {
	inCluster           bool
	serviceAccountEmail string
	subjectEmail        string
}

func NewFromConfig(cfg *config.Config) Builder {
	return Builder{
		inCluster:           cfg.InCluster,
		serviceAccountEmail: fmt.Sprintf("console@%s.iam.gserviceaccount.com", cfg.GoogleManagementProjectID),
		subjectEmail:        "nais-console@" + cfg.TenantDomain,
	}
}

func (g Builder) inClusterTokenSource(ctx context.Context, delegate bool, scopes []string) (oauth2.TokenSource, error) {
	params := google.CredentialsParams{
		Scopes: scopes,
	}
	if delegate {
		params.Subject = g.subjectEmail
	}
	creds, err := google.FindDefaultCredentialsWithParams(ctx, params)
	if err != nil {
		return nil, err
	}

	return creds.TokenSource, nil
}

func (g Builder) impersonateTokenSource(ctx context.Context, delegate bool, scopes []string) (oauth2.TokenSource, error) {
	impersonateConfig := impersonate.CredentialsConfig{
		TargetPrincipal: g.serviceAccountEmail,
		Scopes:          scopes,
	}
	if delegate {
		impersonateConfig.Subject = g.subjectEmail
	}

	return impersonate.CredentialsTokenSource(ctx, impersonateConfig)
}

func (g Builder) client(ctx context.Context, delegate bool, scopes []string) (oauth2.TokenSource, error) {
	if g.inCluster {
		// Use workload identity in cluster
		return g.inClusterTokenSource(ctx, delegate, scopes)
	} else {
		// Otherwise impersonate the service account in project
		return g.impersonateTokenSource(ctx, delegate, scopes)
	}
}

func (g Builder) Admin(ctx context.Context) (oauth2.TokenSource, error) {
	return g.client(ctx, true, []string{
		admin_directory.AdminDirectoryUserReadonlyScope,
		admin_directory.AdminDirectoryGroupScope,
	})
}

func (g Builder) GCP(ctx context.Context) (oauth2.TokenSource, error) {
	return g.client(ctx, false, []string{
		cloudresourcemanager.CloudPlatformScope,
	})
}
