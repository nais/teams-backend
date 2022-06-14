package nais_deploy_reconciler_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	"github.com/stretchr/testify/assert"
)

func TestNaisDeployReconciler_Reconcile(t *testing.T) {
	const teamName = "myteam"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestData := &nais_deploy_reconciler.ProvisionApiKeyRequest{}
		err := json.NewDecoder(r.Body).Decode(requestData)
		assert.NoError(t, err)

		// signed with hmac signature based on timestamp and request data
		assert.Len(t, r.Header.Get("x-nais-signature"), 64)
		assert.Equal(t, teamName, requestData.Team)
		assert.Equal(t, false, requestData.Rotate)

		w.WriteHeader(http.StatusCreated)
	}))

	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	key := make([]byte, 32)
	db := test.GetTestDB()
	logger := auditlogger.New(db)

	systemID, _ := uuid.NewUUID()
	system := dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	reconciler := nais_deploy_reconciler.New(system, logger, http.DefaultClient, srv.URL, key)
	slug := dbmodels.Slug(teamName)

	syncID, _ := uuid.NewUUID()
	corr := dbmodels.Correlation{
		Model: dbmodels.Model{
			ID: &syncID,
		},
	}

	err := reconciler.Reconcile(ctx, reconcilers.NewReconcilerInput(corr, dbmodels.Team{
		Slug: slug,
	}))

	assert.NoError(t, err)
}
