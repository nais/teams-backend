//go:build manual_legacy_test

package google_workspace_admin_reconciler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/nais/console/pkg/test"

	"github.com/nais/console/pkg/slug"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"

	"google.golang.org/api/option"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
)

func TestReconcile(t *testing.T) {
	const (
		domain   = "example.com"
		teamName = "my-team"
	)

	correlationID := uuid.New()
	teamID := uuid.New()
	teamSlug := slug.Slug(teamName)
	team := db.Team{Team: &sqlc.Team{ID: teamID, Name: teamName, Slug: teamSlug, Purpose: sql.NullString{
		String: "some purpose",
		Valid:  true,
	}}}
	teamMembers := []*db.User{
		{
			ID:    uuid.New(),
			Name:  "Some User",
			Email: "mail1@example.com",
		},
		{
			ID:    uuid.New(),
			Name:  "Some Other User",
			Email: "mail2@example.com",
		},
	}

	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers:   teamMembers,
	}

	t.Run("error when unable to load state", func(t *testing.T) {
		ts := test.HttpServerWithHandlers(t, []http.HandlerFunc{})
		defer ts.Close()

		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamID, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		service, _ := admin_directory_v1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, service)
		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "unable to load system state")
	})

	t.Run("empty state, create group", func(t *testing.T) {
		ts := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// create group
			func(w http.ResponseWriter, r *http.Request) {
				grp := admin_directory_v1.Group{}
				_ = json.NewDecoder(r.Body).Decode(&grp)
				assert.Equal(t, "nais-team-my-team@example.com", grp.Email)

				grp.Id = uuid.New().String()
				rsp, _ := grp.MarshalJSON()
				w.Write(rsp)
			},

			// list existing members
			func(w http.ResponseWriter, r *http.Request) {
				// don't set a response
			},
		})
		defer ts.Close()

		ts2 := test.HttpServerWithHandlers(t, []http.HandlerFunc{})
		defer ts2.Close()

		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.On("Logf", ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
			return targets[0].Identifier == teamName
		}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
			return fields.CorrelationID == correlationID
		}), mock.MatchedBy(func(msg string) bool {
			return strings.HasPrefix(msg, "created Google")
		}), "nais-team-my-team@example.com").Return(nil).Once()
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamID, mock.MatchedBy(func(state reconcilers.GoogleWorkspaceState) bool {
				return *state.GroupEmail == "nais-team-my-team@example.com"
			})).
			Return(nil).
			Once()
		service, _ := admin_directory_v1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts2.URL))

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, service)
		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "unable to load system state")
	})
}
