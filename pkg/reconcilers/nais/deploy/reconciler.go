package nais_deploy_reconciler

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
)

type naisDeployReconciler struct {
	database     db.Database
	client       *http.Client
	auditLogger  auditlogger.AuditLogger
	endpoint     string
	provisionKey []byte
	log          logger.Logger
}

const (
	Name              = sqlc.ReconcilerNameNaisDeploy
	metricsSystemName = "nais-deploy"
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, client *http.Client, endpoint string, provisionKey []byte, log logger.Logger) *naisDeployReconciler {
	return &naisDeployReconciler{
		database:     database,
		client:       client,
		auditLogger:  auditLogger,
		endpoint:     endpoint,
		provisionKey: provisionKey,
		log:          log.WithComponent(types.ComponentNameNaisDeploy),
	}
}

func NewFromConfig(_ context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
	provisionKey, err := hex.DecodeString(cfg.NaisDeploy.ProvisionKey)
	if err != nil {
		return nil, err
	}

	return New(database, auditlogger.New(database, types.ComponentNameNaisDeploy, log), http.DefaultClient, cfg.NaisDeploy.Endpoint, provisionKey, log), nil
}

func (r *naisDeployReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *naisDeployReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	payload, err := getProvisionPayload(input.Team.Slug)
	if err != nil {
		return fmt.Errorf("create JSON payload for deploy key API: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request for deploy key API: %w", err)
	}

	signature := genMAC(payload, r.provisionKey)
	request.Header.Set("X-NAIS-Signature", signature)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(request)
	metrics.IncExternalHTTPCalls(metricsSystemName, response, err)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusCreated:
		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionNaisDeployProvisionDeployKey,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Provisioned NAIS deploy API key for team %q", input.Team.Slug)

		now := time.Now()
		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, &reconcilers.NaisDeployKeyState{
			Provisioned: &now,
		})
		if err != nil {
			r.log.WithError(err).Error("persist reconciler state")
		}
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("provision NAIS deploy API key for team %q: %s", input.Team.Slug, response.Status)
	}
}

func (r *naisDeployReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	return nil
}
