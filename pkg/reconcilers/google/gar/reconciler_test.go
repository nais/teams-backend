package gar_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/gar"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/option"
)

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	const (
		managementProjectID = "management-project-123"
		slug                = slug.Slug("team")
	)

	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: slug}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}
	repositoryName := fmt.Sprintf("projects/%s/locations/europe-north1/repositories/%s", managementProjectID, string(slug))
	repositoryDescription := fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", string(slug))

	t.Run("fail when get repository returns error other than 404", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		srv := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "test")
			},
		})
		defer srv.Close()

		b, err := artifactregistry.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		assert.NoError(t, err)

		svc := artifactregistry.NewProjectsLocationsRepositoriesService(b)

		reconciler := gar.New(auditLogger, managementProjectID, svc, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "googleapi: got HTTP response code 500 with body: test")
	})

	t.Run("create when no repository already exists", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		srv := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// Get existing repository
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			// Create call
			func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Path, managementProjectID)

				var repo artifactregistry.Repository
				err := json.NewDecoder(r.Body).Decode(&repo)
				assert.NoError(t, err)
				assert.Equal(t, repositoryName, repo.Name)
				assert.Equal(t, repositoryDescription, repo.Description)
				assert.Equal(t, gar.RepositoryFormat, repo.Format)
				assert.Equal(t, string(slug), repo.Labels["team"])

				response, err := repo.MarshalJSON()
				assert.NoError(t, err)
				op := artifactregistry.Operation{
					Done:     true,
					Response: response,
				}
				resp, err := op.MarshalJSON()
				assert.NoError(t, err)
				w.Write(resp)
			},
		})
		defer srv.Close()

		b, err := artifactregistry.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		assert.NoError(t, err)

		svc := artifactregistry.NewProjectsLocationsRepositoriesService(b)

		reconciler := gar.New(auditLogger, managementProjectID, svc, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("no patch when existing repository has correct values", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		srv := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// Get existing repository
			func(w http.ResponseWriter, r *http.Request) {
				template := &artifactregistry.Repository{
					Format:      gar.RepositoryFormat,
					Name:        repositoryName,
					Description: repositoryDescription,
					Labels: map[string]string{
						"team": string(slug),
					},
				}
				err := json.NewEncoder(w).Encode(template)
				assert.NoError(t, err)

				w.WriteHeader(http.StatusNotFound)
			},
		})
		defer srv.Close()

		b, err := artifactregistry.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		assert.NoError(t, err)

		svc := artifactregistry.NewProjectsLocationsRepositoriesService(b)

		reconciler := gar.New(auditLogger, managementProjectID, svc, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("patch when existing repository has incorrect desc", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		existing := &artifactregistry.Repository{
			Format:      gar.RepositoryFormat,
			Name:        repositoryName,
			Description: repositoryDescription + "invalid",
			Labels: map[string]string{
				"team": string(slug),
			},
		}
		srv := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// Get existing repository
			func(w http.ResponseWriter, r *http.Request) {
				err := json.NewEncoder(w).Encode(existing)
				assert.NoError(t, err)

				w.WriteHeader(http.StatusNotFound)
			},
			// patch repo
			func(w http.ResponseWriter, r *http.Request) {
				var repo artifactregistry.Repository
				err := json.NewDecoder(r.Body).Decode(&repo)
				assert.NoError(t, err)
				assert.Equal(t, existing.Name, repo.Name)
				assert.Equal(t, repositoryDescription, repo.Description)
				assert.Equal(t, existing.Labels, repo.Labels)
				assert.Equal(t, existing.Format, repo.Format)

				response, err := repo.MarshalJSON()
				assert.NoError(t, err)
				op := artifactregistry.Operation{
					Done:     true,
					Response: response,
				}
				resp, err := op.MarshalJSON()
				assert.NoError(t, err)
				w.Write(resp)
			},
		})
		defer srv.Close()

		b, err := artifactregistry.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		assert.NoError(t, err)

		svc := artifactregistry.NewProjectsLocationsRepositoriesService(b)

		reconciler := gar.New(auditLogger, managementProjectID, svc, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}
