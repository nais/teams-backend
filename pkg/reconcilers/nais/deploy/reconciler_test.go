package nais_deploy_reconciler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	nais_deploy_reconciler "github.com/nais/teams-backend/pkg/reconcilers/nais/deploy"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNaisDeployReconciler_Reconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	teamSlug := slug.Slug("slug")
	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: teamSlug}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	key := make([]byte, 32)
	ctx := context.Background()

	t.Run("key successfully provisioned", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestData := &nais_deploy_reconciler.ProvisionApiKeyRequest{}
			err := json.NewDecoder(r.Body).Decode(requestData)
			assert.NoError(t, err)
			assert.Len(t, r.Header.Get("x-nais-signature"), 64)
			assert.Equal(t, string(teamSlug), requestData.Team)
			assert.Equal(t, false, requestData.Rotate)

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameNaisDeploy).
			Return(auditLogger).
			Once()

		database := db.NewMockDatabase(t)

		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return t[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionNaisDeployProvisionDeployKey && f.CorrelationID == correlationID
			}), mock.Anything, mock.Anything).
			Return(nil).
			Once()

		database.
			On("SetReconcilerStateForTeam", ctx, sqlc.ReconcilerNameNaisDeploy, teamSlug, mock.Anything).
			Return(nil).
			Once()

		err := nais_deploy_reconciler.
			New(database, auditLogger, http.DefaultClient, srv.URL, key, log).
			Reconcile(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("internal server error when provisioning key", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameNaisDeploy).
			Return(auditLogger).
			Once()

		database := db.NewMockDatabase(t)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		err := nais_deploy_reconciler.
			New(database, auditLogger, http.DefaultClient, srv.URL, key, log).
			Reconcile(ctx, input)
		assert.EqualError(t, err, "provision NAIS deploy API key for team \"slug\": 500 Internal Server Error")
	})

	t.Run("team key does not change", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameNaisDeploy).
			Return(auditLogger).
			Once()

		database := db.NewMockDatabase(t)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := nais_deploy_reconciler.
			New(database, auditLogger, http.DefaultClient, srv.URL, key, log).
			Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}
