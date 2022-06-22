package nais_deploy_reconciler_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNaisDeployReconciler_Reconcile(t *testing.T) {
	const (
		teamName = "My Team"
		teamSlug = "slug"
	)

	system := dbmodels.System{Model: modelWithId()}
	corr := dbmodels.Correlation{Model: modelWithId()}
	team := dbmodels.Team{Model: modelWithId(), Name: teamName, Slug: teamSlug}
	key := make([]byte, 32)
	auditLogger := &auditlogger.MockAuditLogger{}
	input := reconcilers.Input{
		Corr: corr,
		Team: team,
	}

	t.Run("key successfully provisioned", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestData := &nais_deploy_reconciler.ProvisionApiKeyRequest{}
			err := json.NewDecoder(r.Body).Decode(requestData)
			assert.NoError(t, err)

			// signed with hmac signature based on timestamp and request data
			assert.Len(t, r.Header.Get("x-nais-signature"), 64)
			assert.Equal(t, teamSlug, requestData.Team)
			assert.Equal(t, false, requestData.Rotate)

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		auditLogger.
			On(
				"Logf",
				nais_deploy_reconciler.OpProvisionApiKey,
				corr,
				system,
				mock.Anything,
				&team,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).
			Return(nil).
			Once()

		reconciler := nais_deploy_reconciler.New(system, auditLogger, http.DefaultClient, srv.URL, key)
		err := reconciler.Reconcile(context.Background(), input)

		assert.NoError(t, err)
	})

	t.Run("internal server error when provisioning key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		reconciler := nais_deploy_reconciler.New(system, auditLogger, http.DefaultClient, srv.URL, key)
		err := reconciler.Reconcile(context.Background(), input)
		assert.EqualError(t, err, "provision NAIS deploy API key to team 'slug': 500 Internal Server Error")
	})

	t.Run("team key does not change", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()
		reconciler := nais_deploy_reconciler.New(system, auditLogger, http.DefaultClient, srv.URL, key)
		err := reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})
}

func modelWithId() dbmodels.Model {
	id, _ := uuid.NewUUID()
	return dbmodels.Model{ID: &id}
}
